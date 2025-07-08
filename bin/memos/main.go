package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/pixb/memos-server/server"
	"github.com/pixb/memos-server/server/profile"
	"github.com/pixb/memos-server/server/version"
	"github.com/pixb/memos-server/store"
	"github.com/pixb/memos-server/store/db"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	greetingBanner = `
███╗   ███╗███████╗███╗   ███╗ ██████╗ ███████╗
████╗ ████║██╔════╝████╗ ████║██╔═══██╗██╔════╝
██╔████╔██║█████╗  ██╔████╔██║██║   ██║███████╗
██║╚██╔╝██║██╔══╝  ██║╚██╔╝██║██║   ██║╚════██║
██║ ╚═╝ ██║███████╗██║ ╚═╝ ██║╚██████╔╝███████║
╚═╝     ╚═╝╚══════╝╚═╝     ╚═╝ ╚═════╝ ╚══════╝
`
)

/*
- Use: 命令名称
- Short: 命令帮助显示描述。
- 创建配置实例[[profile.go]]
- [[viper]]获取配置。
- [[version.go]]获取当前版本。
- 验证配置。
- 拷贝一个可以取消的`Context`: [[go_context#context.WithCancel()]]
- 创建数据库驱动[[db.go]]
- 创建存储实例[[store.go]]
- 创建服务实例[[work/memos/memos-代码级总结/server.go|server.go]]
- 优雅关机逻辑[[go-优雅的关机或重启]]
- 启动服务端服务。
- 打印欢迎信息
- 异步线程处理优雅关机消息。
- 等待`Ctrl + c`中止。
*/
var rootCmd = &cobra.Command{
	Use:   "memos",
	Short: `An open source, lightweight note-taking service. Easily capture and share your great thoughts.`,
	Run: func(_ *cobra.Command, _ []string) {
		instanceProfile := &profile.Profile{
			Mode:        viper.GetString("mode"),
			Addr:        viper.GetString("addr"),
			Port:        viper.GetInt("port"),
			UNIXSock:    viper.GetString("unix-sock"),
			Data:        viper.GetString("data"),
			Driver:      viper.GetString("driver"),
			DSN:         viper.GetString("dsn"),
			InstanceURL: viper.GetString("instance-url"),
			Version:     version.GetCurrentVersion(viper.GetString("mode")),
		}
		if err := instanceProfile.Validate(); err != nil {
			panic(err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		log.Printf("db.NewDBDriver(instanceProfile), instanceProfile:%v\n", instanceProfile)
		dbDriver, err := db.NewDBDriver(instanceProfile)
		if err != nil {
			cancel()
			slog.Error("failed to create db driver", "error", err)
			return
		}

		log.Println("store.New(dbDriver, instanceProfile)")
		storeInstance := store.New(dbDriver, instanceProfile)
		if err := storeInstance.Migrate(ctx); err != nil {
			cancel()
			slog.Error("failed to migrate", "error", err)
			return
		}

		log.Println("server.NewServer(ctx, instanceProfile, storeInstance)")
		s, err := server.NewServer(ctx, instanceProfile, storeInstance)
		if err != nil {
			cancel()
			slog.Error("failed to create server", "error", err)
			return
		}

		c := make(chan os.Signal, 1)
		// Trigger graceful shutdown on SIGINT or SIGTERM.
		// The default signal sent by the `kill` command is SIGTERM,
		// which is taken as the graceful shutdown signal for many systems, eg., Kubernetes, Gunicorn.
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)

		log.Println("s.Start(ctx)")
		if err := s.Start(ctx); err != nil {
			if err != http.ErrServerClosed {
				slog.Error("failed to start server", "error", err)
				cancel()
			}
		}
		//
		printGreetings(instanceProfile)

		go func() {
			<-c
			log.Println("s.Shutdown(ctx)")
			// s.Shutdown(ctx)
			cancel()
		}()

		// Wait for CTRL-C.
		<-ctx.Done()
	},
}

/**
* 初始化
* - 设置viper默认值
* 	- `mode`: 开发模式dev
* 	- `driver`: 数据库驱动sqlite
* 	- `port`: 端口8081
* - 定义[[cobra]]的全局命令参数
* - 绑定`pflag`,根据名称从`cobra`全局配置中查找，返回`Flag`对象，绑定到viper中。
* - 设置环境变量前缀
* - 自动将环境变量绑定到viper
**/
func init() {
	viper.SetDefault("mode", "dev")
	viper.SetDefault("driver", "sqlite")
	viper.SetDefault("port", 8081)

	rootCmd.PersistentFlags().String("mode", "dev", `mode of server, can be "prod" or "dev" or "demo"`)
	rootCmd.PersistentFlags().String("addr", "", "address of server")
	rootCmd.PersistentFlags().Int("port", 8081, "port of server")
	rootCmd.PersistentFlags().String("unix-sock", "", "path to the unix socket, overrides --addr and --port")
	rootCmd.PersistentFlags().String("data", "", "data directory")
	rootCmd.PersistentFlags().String("driver", "sqlite", "database driver")
	rootCmd.PersistentFlags().String("dsn", "", "database source name(aka. DSN)")
	rootCmd.PersistentFlags().String("instance-url", "", "the url of your memos instance")

	if err := viper.BindPFlag("mode", rootCmd.PersistentFlags().Lookup("mode")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("addr", rootCmd.PersistentFlags().Lookup("addr")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("unix-sock", rootCmd.PersistentFlags().Lookup("unix-sock")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("data", rootCmd.PersistentFlags().Lookup("data")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("driver", rootCmd.PersistentFlags().Lookup("driver")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("dsn", rootCmd.PersistentFlags().Lookup("dsn")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("instance-url", rootCmd.PersistentFlags().Lookup("instance-url")); err != nil {
		panic(err)
	}

	viper.SetEnvPrefix("memos")
	viper.AutomaticEnv()
	if err := viper.BindEnv("instance-url", "MEMOS_INSTANCE_URL"); err != nil {
		panic(err)
	}
}

// 打印欢迎消息
func printGreetings(profile *profile.Profile) {
	if profile.IsDev() {
		println("Development mode is enabled")
		println("DSN: ", profile.DSN)
	}
	fmt.Printf(`---
Server profile
version: %s
data: %s
addr: %s
port: %d
unix-sock: %s
mode: %s
driver: %s
---
`, profile.Version, profile.Data, profile.Addr, profile.Port, profile.UNIXSock, profile.Mode, profile.Driver)

	print(greetingBanner)
	if len(profile.UNIXSock) == 0 {
		if len(profile.Addr) == 0 {
			fmt.Printf("Version %s has been started on port %d\n", profile.Version, profile.Port)
		} else {
			fmt.Printf("Version %s has been started on address '%s' and port %d\n", profile.Version, profile.Addr, profile.Port)
		}
	} else {
		fmt.Printf("Version %s has been started on unix socket %s\n", profile.Version, profile.UNIXSock)
	}
	fmt.Printf(`---
See more in:
👉Website: %s
👉GitHub: %s
---
`, "https://usememos.com", "https://github.com/usememos/memos")
}

func main() {
	// cobra Command Run
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
