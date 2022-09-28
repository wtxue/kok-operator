package commands

import (
	"fmt"
	"io"
	"os"

	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/shiena/ansicolor"
	"github.com/spf13/pflag"

	"github.com/wtxue/kok-operator/pkg/apis"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2/klogr"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	colorErr = color.RedString("Error")

	// errAction specifies what should happen when an error occurs
	errAction = func() {
		os.Exit(1)
	}
)

func init() {
	color.Output = ansicolor.NewAnsiColorWriter(os.Stderr)
}

// GlobalManagerOption ...
type GlobalManagerOption struct {
	Kubeconfig       string
	Context          string
	Trace            bool
	Verbose          bool
	EnableKlog       bool
	EnableDevLogging bool
	LogLevel         string
	OutputFormat     string
	Writer           io.Writer
}

func defaultGlobalManagerOption() *GlobalManagerOption {
	return &GlobalManagerOption{
		OutputFormat: "text",
		Writer:       color.Output,
		// Writer:       os.Stdout,
	}

}

// AddFlags adds flags for a specific server to the specified FlagSet object.
func (o *GlobalManagerOption) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.Kubeconfig, "kubeconfig", "c", "", "path to the kubeconfig file to use for vkectl requests")
	fs.StringVar(&o.Context, "context", "", "name of the kubeconfig context to use")
	fs.BoolVarP(&o.Verbose, "verbose", "v", false, "turn on debug logging")
	fs.StringVarP(&o.OutputFormat, "output", "o", o.OutputFormat, "output format (text|json)")
}

// SetupLogger initializes the logger used in the service controller
func (o *GlobalManagerOption) SetupLogger() {
	if o.EnableKlog {
		logf.SetLogger(klogr.New())
	} else {
		var lvl zapcore.LevelEnabler

		switch o.LogLevel {
		case "debug":
			lvl = zapcore.DebugLevel
		default:
			lvl = zapcore.InfoLevel
		}

		zapOptions := zap.Options{
			Development: o.EnableDevLogging,
			Level:       lvl,
		}
		logf.SetLogger(zap.New(zap.UseFlagOptions(&zapOptions)))
	}
}

// GetClientConfigWithContext ...
func GetClientConfigWithContext(kubeconfigPath, kubeContext string) clientcmd.ClientConfig {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	if kubeconfigPath != "" {
		rules.ExplicitPath = kubeconfigPath
	}
	overrides := &clientcmd.ConfigOverrides{CurrentContext: kubeContext}
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides)
}

// GetConfigWithContext GetConfig returns kubernetes config based on the current environment.
// If path is provided, loads configuration from that file. Otherwise,
// GetConfig uses default strategy to load configuration from $KUBECONFIG,
// .kube/config, or just returns in-cluster config.
func GetConfigWithContext(kubeconfigPath, kubeContext string) (*rest.Config, error) {
	return GetClientConfigWithContext(kubeconfigPath, kubeContext).ClientConfig()
}

// GetK8sConfig defines the function for get k8s config.
func (o *GlobalManagerOption) GetK8sConfig() (*rest.Config, error) {
	cfg, err := GetConfigWithContext(o.Kubeconfig, "")
	if err != nil {
		return nil, errors.Wrap(err, "could not get k8s config")
	}

	return cfg, nil
}

func checkErr(err error) {
	if err == nil {
		return
	}

	fmt.Fprintf(color.Output, "%s: %v\n", colorErr, err)
	errAction()
}

func info(msg string, args ...interface{}) {
	fmt.Fprintf(color.Output, "%s\n", color.YellowString(msg, args...))
}

// marshalToJSONForCodecs ...
func marshalToJSONForCodecs(obj runtime.Object, gv schema.GroupVersion, codecs serializer.CodecFactory) ([]byte, error) {
	const mediaType = runtime.ContentTypeJSON
	info, ok := runtime.SerializerInfoForMediaType(codecs.SupportedMediaTypes(), mediaType)
	if !ok {
		return []byte{}, errors.Errorf("unsupported media type %q", mediaType)
	}

	encoder := codecs.EncoderForVersion(info.Serializer, gv)
	return runtime.Encode(encoder, obj)
}

// MarshalToJSON marshals an object into json.
func MarshalToJSON(obj runtime.Object) ([]byte, error) {
	gvk, _ := apiutil.GVKForObject(obj, apis.GetScheme())
	obj.GetObjectKind().SetGroupVersionKind(gvk)

	return marshalToJSONForCodecs(obj, gvk.GroupVersion(), serializer.NewCodecFactory(apis.GetScheme()))
}

// Serializer ...
var Serializer = json.NewSerializerWithOptions(json.DefaultMetaFactory,
	GetScheme(), GetScheme(), json.SerializerOptions{Yaml: false})
