module github.com/golangci/golangci-lint

go 1.12

require (
	github.com/OpenPeeDeeP/depguard v1.0.1
	github.com/StackExchange/wmi v0.0.0-20190523213315-cbe66965904d // indirect
	github.com/fatih/color v1.7.0
	github.com/go-critic/go-critic v0.3.5-0.20190904082202-d79a9f0c64db
	github.com/go-lintpack/lintpack v0.5.2
	github.com/go-ole/go-ole v1.2.4 // indirect
	github.com/gogo/protobuf v1.3.0 // indirect
	github.com/golangci/check v0.0.0-20180506172741-cfe4005ccda2
	github.com/golangci/dupl v0.0.0-20180902072040-3e9179ac440a
	github.com/golangci/errcheck v0.0.0-20181223084120-ef45e06d44b6
	github.com/golangci/go-misc v0.0.0-20180628070357-927a3d87b613
	github.com/golangci/goconst v0.0.0-20180610141641-041c5f2b40f3
	github.com/golangci/gocyclo v0.0.0-20180528144436-0a533e8fa43d
	github.com/golangci/gofmt v0.0.0-20190930125516-244bba706f1a
	github.com/golangci/ineffassign v0.0.0-20190609212857-42439a7714cc
	github.com/golangci/lint-1 v0.0.0-20190420132249-ee948d087217
	github.com/golangci/maligned v0.0.0-20180506175553-b1d89398deca
	github.com/golangci/misspell v0.0.0-20180809174111-950f5d19e770
	github.com/golangci/prealloc v0.0.0-20180630174525-215b22d4de21
	github.com/golangci/revgrep v0.0.0-20180812185044-276a5c0a1039
	github.com/golangci/unconvert v0.0.0-20180507085042-28b1c447d1f4
	github.com/google/go-cmp v0.3.1 // indirect
	github.com/gostaticanalysis/analysisutil v0.0.3 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/magiconair/properties v1.8.1 // indirect
	github.com/matoous/godox v0.0.0-20190911065817-5d6d842e92eb
	github.com/mattn/go-colorable v0.1.4
	github.com/mattn/go-isatty v0.0.9 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/go-ps v0.0.0-20190716172923-621e5597135b
	github.com/onsi/ginkgo v1.10.2 // indirect
	github.com/pelletier/go-toml v1.4.0 // indirect
	github.com/pkg/errors v0.8.1
	github.com/securego/gosec v0.0.0-20191002120514-e680875ea14d
	github.com/shirou/gopsutil v2.19.9+incompatible // v2.19.8
	github.com/shirou/w32 v0.0.0-20160930032740-bb4de0191aa4 // indirect
	github.com/shurcooL/go v0.0.0-20190704215121-7189cc372560 // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/sourcegraph/go-diff v0.5.1
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/cobra v0.0.5
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.4.0
	github.com/timakin/bodyclose v0.0.0-20190930140734-f7f2e9bca95e
	github.com/ultraware/funlen v0.0.2
	github.com/ultraware/whitespace v0.0.3
	github.com/valyala/quicktemplate v1.2.0
	golang.org/x/net v0.0.0-20191002035440-2ec189313ef0 // indirect
	golang.org/x/sys v0.0.0-20191002091554-b397fe3ad8ed // indirect
	golang.org/x/tools v0.0.0-20191002234911-9ade4c73f2af
	gopkg.in/yaml.v2 v2.2.4
	honnef.co/go/tools v0.0.1-2019.2.3
	mvdan.cc/interfacer v0.0.0-20180901003855-c20040233aed
	mvdan.cc/lint v0.0.0-20170908181259-adc824a0674b // indirect
	mvdan.cc/unparam v0.0.0-20190917161559-b83a221c10a2
	sourcegraph.com/sqs/pbtypes v1.0.0 // indirect
)

// https://github.com/golang/tools/pull/156
// https://github.com/golang/tools/pull/160
// https://github.com/golang/tools/pull/162
replace golang.org/x/tools => github.com/golangci/tools v0.0.0-20190915081525-6aa350649b1c
