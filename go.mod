module github.com/cwbudde/matplotlib-go

go 1.25.0

require (
	github.com/cwbudde/algo-fft v0.0.0
	github.com/spf13/cobra v1.9.1
	github.com/spf13/viper v1.20.1
	golang.org/x/image v0.30.0 // pinned to commit c574db581976698ac047466629eeeb7b17bb49dd for determinism
)

require (
	codeberg.org/go-fonts/dejavu v0.4.0
	gioui.org v0.10.0
	github.com/cwbudde/agg_go v0.2.31
	golang.org/x/text v0.32.0
)

require (
	gioui.org/shader v1.0.8 // indirect
	github.com/fsnotify/fsnotify v1.8.0 // indirect
	github.com/go-text/typesetting v0.3.4 // indirect
	github.com/go-viper/mapstructure/v2 v2.2.1 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.3 // indirect
	github.com/sagikazarmark/locafero v0.7.0 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.12.0 // indirect
	github.com/spf13/cast v1.7.1 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.9.0 // indirect
	golang.org/x/exp/shiny v0.0.0-20250408133849-7e4ce0ab07d0 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/cwbudde/algo-fft => ../algo-fft

replace github.com/cwbudde/agg_go => ../agg_go
