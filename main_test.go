package main

import (
	"fmt"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestParseYAML(t *testing.T) {
	Convey("parseYAML()", t, func() {
		Convey("returns error for invalid yaml data", func() {
			data := `
asdf: fdsa
- asdf: fdsa
`
			obj, err := parseYAML([]byte(data))
			So(err.Error(), ShouldStartWith, "unmarshal []byte to yaml failed:")
			So(obj, ShouldBeNil)
		})
		Convey("returns error if yaml was not a top level map", func() {
			data := `
- 1
- 2
`
			obj, err := parseYAML([]byte(data))
			So(err.Error(), ShouldStartWith, "Root of YAML document is not a hash/map:")
			So(obj, ShouldBeNil)
		})
		Convey("returns expected datastructure from valid yaml", func() {
			data := `
top:
  subarray:
  - one
  - two
`
			obj, err := parseYAML([]byte(data))
			expect := map[interface{}]interface{}{
				"top": map[interface{}]interface{}{
					"subarray": []interface{}{"one", "two"},
				},
			}
			So(obj, ShouldResemble, expect)
			So(err, ShouldBeNil)
		})
	})
}

func TestMergeAllDocs(t *testing.T) {
	Convey("mergeAllDocs()", t, func() {
		Convey("Fails with readFile error on bad first doc", func() {
			target := map[interface{}]interface{}{}
			err := mergeAllDocs(target, []string{"assets/merge/nonexistent.yml", "assets/merge/second.yml"})
			So(err.Error(), ShouldStartWith, "Error reading file assets/merge/nonexistent.yml:")
		})
		Convey("Fails with parseYAML error on bad second doc", func() {
			target := map[interface{}]interface{}{}
			err := mergeAllDocs(target, []string{"assets/merge/first.yml", "assets/merge/bad.yml"})
			So(err.Error(), ShouldStartWith, "assets/merge/bad.yml: Root of YAML document is not a hash/map:")
		})
		Convey("Fails with mergeMap error", func() {
			target := map[interface{}]interface{}{}
			err := mergeAllDocs(target, []string{"assets/merge/first.yml", "assets/merge/error.yml"})
			So(err.Error(), ShouldStartWith, "assets/merge/error.yml: $.array_inline.0: new object is a string, not a map - cannot merge using keys")
		})
		Convey("Succeeds with valid files + yaml", func() {
			target := map[interface{}]interface{}{}
			expect := map[interface{}]interface{}{
				"key":           "overridden",
				"array_append":  []interface{}{"one", "two", "three"},
				"array_prepend": []interface{}{"three", "four", "five"},
				"array_inline": []interface{}{
					map[interface{}]interface{}{"name": "first_elem", "val": "overwritten"},
					"second_elem was overwritten",
					"third elem is appended",
				},
				"map": map[interface{}]interface{}{
					"key":  "value",
					"key2": "val2",
				},
			}
			err := mergeAllDocs(target, []string{"assets/merge/first.yml", "assets/merge/second.yml"})
			So(err, ShouldBeNil)
			So(target, ShouldResemble, expect)
		})
	})
}

func TestMain(t *testing.T) {
	Convey("main()", t, func() {
		var stdout string
		printfStdOut = func(format string, args ...interface{}) {
			stdout = fmt.Sprintf(format, args...)
		}
		var stderr string
		printfStdErr = func(format string, args ...interface{}) {
			stderr = fmt.Sprintf(format, args...)
		}

		rc := 256 // invalid return code to catch any issues
		exit = func(code int) {
			rc = code
		}

		usage = func() {
			stderr = "usage was called"
			exit(1)
		}

		Convey("Should output usage if bad args are passed", func() {
			os.Args = []string{"spruce", "fdsafdada"}
			stdout = ""
			stderr = ""
			main()
			So(stderr, ShouldEqual, "usage was called")
			So(rc, ShouldEqual, 1)
		})
		Convey("Should output usage if no args at all", func() {
			os.Args = []string{"spruce"}
			stdout = ""
			stderr = ""
			main()
			So(stderr, ShouldEqual, "usage was called")
			So(rc, ShouldEqual, 1)
		})
		Convey("Should output usage if no args to merge", func() {
			os.Args = []string{"spruce", "merge"}
			stdout = ""
			stderr = ""
			main()
			So(stderr, ShouldEqual, "usage was called")
			So(rc, ShouldEqual, 1)
		})
		Convey("Should output version", func() {
			Convey("When '-v' is specified", func() {
				os.Args = []string{"spruce", "-v"}
				stdout = ""
				stderr = ""
				main()
				So(stdout, ShouldEqual, "")
				So(stderr, ShouldEqual, fmt.Sprintf("spruce - Version %s\n", VERSION))
				So(rc, ShouldEqual, 0)
			})
			Convey("When '--version' is specified", func() {
				os.Args = []string{"spruce", "--version"}
				stdout = ""
				stderr = ""
				main()
				So(stdout, ShouldEqual, "")
				So(stderr, ShouldEqual, fmt.Sprintf("spruce - Version %s\n", VERSION))
				So(rc, ShouldEqual, 0)
			})
		})
		Convey("Should panic on errors merging docs", func() {
			os.Args = []string{"spruce", "merge", "assets/merge/bad.yml"}
			stdout = ""
			stderr = ""
			main()
			So(stderr, ShouldStartWith, "assets/merge/bad.yml: Root of YAML document is not a hash/map:")
			So(rc, ShouldEqual, 2)
		})
		/* Fixme - how to trigger this?
		Convey("Should panic on errors marshalling yaml", func () {
		})
		*/
		Convey("Should output merged yaml on success", func() {
			os.Args = []string{"spruce", "merge", "assets/merge/first.yml", "assets/merge/second.yml"}
			stdout = ""
			stderr = ""
			main()
			So(stdout, ShouldEqual, `array_append:
- one
- two
- three
array_inline:
- name: first_elem
  val: overwritten
- second_elem was overwritten
- third elem is appended
array_prepend:
- three
- four
- five
key: overridden
map:
  key: value
  key2: val2

`)
			So(stderr, ShouldEqual, "")
		})
		Convey("Should not fail when handling concourse-style yaml and --concourse", func() {
			os.Args = []string{"spruce", "--concourse", "merge", "assets/concourse/first.yml", "assets/concourse/second.yml"}
			stdout = ""
			stderr = ""
			main()
			So(stdout, ShouldEqual, `jobs:
- curlies: {{my-variable_123}}
  name: thing1
- curlies: {{more}}
  name: thing2

`)
			So(stderr, ShouldEqual, "")
		})

		Convey("Should handle de-referencing", func() {
			os.Args = []string{"spruce", "merge", "assets/dereference/first.yml", "assets/dereference/second.yml"}
			stdout = ""
			stderr = ""
			main()
			So(stdout, ShouldEqual, `jobs:
- name: my-server
  static_ips:
  - 192.168.1.0
properties:
  client:
    servers:
    - 192.168.1.0

`)
			So(stderr, ShouldEqual, "")
		})
		Convey("De-referencing cyclical datastructures should throw an error", func() {
			os.Args = []string{"spruce", "merge", "assets/dereference/cyclic-data.yml"}
			stdout = ""
			stderr = ""
			main()
			So(stdout, ShouldEqual, "")
			So(stderr, ShouldEndWith, "hit max recursion depth. You seem to have a self-referencing dataset")
			So(rc, ShouldEqual, 2)
		})
		Convey("Dereferencing multiple values should behave as desired", func() {
			os.Args = []string{"spruce", "merge", "assets/dereference/multi-value.yml"}
			stdout = ""
			stderr = ""
			main()
			So(stdout, ShouldEqual, `jobs:
- instances: 1
  name: api_z1
  networks:
  - name: net1
    static_ips:
    - 192.168.1.2
    - 192.168.1.3
    - 192.168.1.4
- instances: 1
  name: api_z2
  networks:
  - name: net2
    static_ips:
    - 192.168.2.2
    - 192.168.2.3
    - 192.168.2.4
networks:
- name: net1
  subnets:
  - cloud_properties: random
    static:
    - 192.168.1.2 - 192.168.1.30
- name: net2
  subnets:
  - cloud_properties: random
    static:
    - 192.168.2.2 - 192.168.2.30
properties:
  api_server_primary: 192.168.1.2
  api_servers:
  - 192.168.1.2
  - 192.168.1.3
  - 192.168.1.4
  - 192.168.2.2
  - 192.168.2.3
  - 192.168.2.4

`)
			So(stderr, ShouldEqual, "")
		})
		Convey("Should output error on bad de-reference", func() {
			os.Args = []string{"spruce", "merge", "assets/dereference/bad.yml"}
			stdout = ""
			stderr = ""
			main()
			So(stderr, ShouldStartWith, "$.bad.dereference: Unable to resolve `my.value`")
			So(rc, ShouldEqual, 2)
		})
		Convey("Pruning should happen after de-referencing", func() {
			os.Args = []string{"spruce", "merge", "--prune", "jobs", "--prune", "properties.client.servers", "assets/dereference/first.yml", "assets/dereference/second.yml"}
			stdout = ""
			stderr = ""
			main()
			So(stderr, ShouldEqual, "")
			So(stdout, ShouldEqual, `properties:
  client: {}

`)
		})
		Convey("static_ips() failures return errors to the user", func() {
			os.Args = []string{"spruce", "merge", "assets/static_ips/jobs.yml"}
			stdout = ""
			stderr = ""
			main()
			So(stderr, ShouldEqual, "$.jobs.api_z1.networks.net1.static_ips: `$.networks` could not be found in the YAML datastructure")
			So(stdout, ShouldEqual, "")
		})
		Convey("static_ips() get resolved, and are resolved prior to dereferencing", func() {
			os.Args = []string{"spruce", "merge", "assets/static_ips/properties.yml", "assets/static_ips/jobs.yml", "assets/static_ips/network.yml"}
			stdout = ""
			stderr = ""
			main()
			So(stderr, ShouldEqual, "")
			So(stdout, ShouldEqual, `jobs:
- instances: 3
  name: api_z1
  networks:
  - name: net1
    static_ips:
    - 10.0.0.2
    - 10.0.0.3
    - 10.0.0.4
networks:
- name: net1
  subnets:
  - static:
    - 10.0.0.2 - 10.0.0.20
properties:
  api_servers:
  - 10.0.0.2
  - 10.0.0.3
  - 10.0.0.4

`)
		})
		Convey("Parameters override their requirement", func() {
			os.Args = []string{"spruce", "merge", "assets/params/global.yml", "assets/params/good.yml"}
			stdout = ""
			stderr = ""
			main()
			So(stdout, ShouldEqual, `cpu: 3
nested:
  key:
    override: true
networks:
- true
storage: 4096

`)
			So(stderr, ShouldEqual, "")
		})
		Convey("Parameters must be specified", func() {
			os.Args = []string{"spruce", "merge", "assets/params/global.yml", "assets/params/fail.yml"}
			stdout = ""
			stderr = ""
			main()
			So(stdout, ShouldEqual, "")
			So(stderr, ShouldEqual, "$.nested.key.override: provide nested override")
		})
		Convey("Pruning takes place before parameters", func() {
			os.Args = []string{"spruce", "merge", "--prune", "nested", "assets/params/global.yml", "assets/params/fail.yml"}
			stdout = ""
			stderr = ""
			main()
			So(stdout, ShouldEqual, `cpu: 3
networks: specified
storage: 4096

`)
			So(stderr, ShouldEqual, "")
		})
	})
}

func TestDebug(t *testing.T) {
	var stderr string
	usage = func() {}
	printfStdErr = func(format string, args ...interface{}) {
		stderr = fmt.Sprintf(format, args...)
	}
	Convey("debug", t, func() {
		Convey("Outputs when debug is set to true", func() {
			stderr = ""
			debug = true
			DEBUG("test debugging")
			So(stderr, ShouldEqual, "DEBUG> test debugging\n")
		})
		Convey("Multi-line debug inputs are each prefixed", func() {
			stderr = ""
			debug = true
			DEBUG("test debugging\nsecond line")
			So(stderr, ShouldEqual, "DEBUG> test debugging\nDEBUG> second line\n")
		})
		Convey("Doesn't output when debug is set to false", func() {
			stderr = ""
			debug = false
			DEBUG("test debugging")
			So(stderr, ShouldEqual, "")
		})
	})
	Convey("debug flags:", t, func() {
		Convey("-D enables debugging", func() {
			os.Args = []string{"spruce", "-D"}
			debug = false
			main()
			So(debug, ShouldBeTrue)
		})
		Convey("--debug enables debugging", func() {
			os.Args = []string{"spruce", "--debug"}
			debug = false
			main()
			So(debug, ShouldBeTrue)
		})
		Convey("DEBUG=\"tRuE\" enables debugging", func() {
			os.Setenv("DEBUG", "tRuE")
			os.Args = []string{"spruce"}
			debug = false
			main()
			So(debug, ShouldBeTrue)
		})
		Convey("DEBUG=1 enables debugging", func() {
			os.Setenv("DEBUG", "1")
			os.Args = []string{"spruce"}
			debug = false
			main()
			So(debug, ShouldBeTrue)
		})
		Convey("DEBUG=randomval enables debugging", func() {
			os.Setenv("DEBUG", "randomval")
			os.Args = []string{"spruce"}
			debug = false
			main()
			So(debug, ShouldBeTrue)
		})
		Convey("DEBUG=\"fAlSe\" disables debugging", func() {
			os.Setenv("DEBUG", "fAlSe")
			os.Args = []string{"spruce"}
			debug = false
			main()
			So(debug, ShouldBeFalse)
		})
		Convey("DEBUG=0 disables debugging", func() {
			os.Setenv("DEBUG", "0")
			os.Args = []string{"spruce"}
			debug = false
			main()
			So(debug, ShouldBeFalse)
		})
		Convey("DEBUG=\"\" disables debugging", func() {
			os.Setenv("DEBUG", "")
			os.Args = []string{"spruce"}
			debug = false
			main()
			So(debug, ShouldBeFalse)
		})
	})
}

func TestQuoteConcourse(t *testing.T) {
	Convey("quoteConcourse()", t, func() {
		Convey("Correctly double-quotes incoming {{\\S}} patterns", func() {
			Convey("adds quotes", func() {
				input := []byte("name: {{var-_1able}}")
				So(string(quoteConcourse(input)), ShouldEqual, "name: \"{{var-_1able}}\"")
			})
		})
		Convey("doesn't affect regularly quoted things", func() {
			input := []byte("name: \"my value\"")
			So(string(quoteConcourse(input)), ShouldEqual, "name: \"my value\"")
		})
	})
}
func TestDequoteConcourse(t *testing.T) {
	Convey("dequoteConcourse()", t, func() {
		Convey("Correctly removes quotes from incoming {{\\S}} patterns", func() {
			Convey("with single quotes", func() {
				input := []byte("name: '{{var-_1able}}'")
				So(dequoteConcourse(input), ShouldEqual, "name: {{var-_1able}}")
			})
			Convey("with double quotes", func() {
				input := []byte("name: \"{{var-_1able}}\"")
				So(dequoteConcourse(input), ShouldEqual, "name: {{var-_1able}}")
			})
		})
		Convey("doesn't affect regularly quoted things", func() {
			input := []byte("name: \"my value\"")
			So(dequoteConcourse(input), ShouldEqual, "name: \"my value\"")
		})
	})
}
