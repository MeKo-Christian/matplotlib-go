module github.com/cwbudde/matplotlib-go

go 1.25.0

require (
	github.com/cwbudde/algo-fft v0.6.11
	github.com/spf13/cobra v1.10.2
	github.com/spf13/viper v1.21.0
	golang.org/x/image v0.42.0 // pinned to commit c574db581976698ac047466629eeeb7b17bb49dd for determinism
)

require (
	codeberg.org/go-fonts/dejavu v0.4.0
	gioui.org v0.10.0
	github.com/cwbudde/agg_go v0.3.2
	github.com/cwbudde/mathtext v0.4.3
	golang.org/x/net v0.56.0
	golang.org/x/text v0.38.0
)

require (
	gioui.org/shader v1.0.8 // indirect
	github.com/MeKo-Christian/qhull-go v0.1.0
	github.com/fsnotify/fsnotify v1.10.1 // indirect
	github.com/go-text/typesetting v0.3.4 // indirect
	github.com/go-viper/mapstructure/v2 v2.5.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/pelletier/go-toml/v2 v2.4.0 // indirect
	github.com/sagikazarmark/locafero v0.12.0 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/exp/shiny v0.0.0-20260611194520-c48552f49976 // indirect
	golang.org/x/sys v0.46.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)

replace github.com/MeKo-Christian/qhull-go => ../qhull-go
