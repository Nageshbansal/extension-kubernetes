package extcommon

import (
	"bytes"
	"github.com/rs/zerolog/log"
	"github.com/zegl/kube-score/config"
	ks "github.com/zegl/kube-score/domain"
	"github.com/zegl/kube-score/parser"
	"github.com/zegl/kube-score/score"
	"github.com/zegl/kube-score/scorecard"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sJson "k8s.io/apimachinery/pkg/runtime/serializer/json"
	"strings"
)

var (
	serializer = k8sJson.NewSerializerWithOptions(
		k8sJson.DefaultMetaFactory, nil, nil,
		k8sJson.SerializerOptions{
			Yaml:   true,
			Pretty: true,
			Strict: true,
		},
	)
)

func getKubeScore(manifest string) (*scorecard.Scorecard, error) {

	reader := &inputReader{
		Reader: strings.NewReader(manifest),
	}

	cnf := config.Configuration{
		AllFiles:                              []ks.NamedReader{reader},
		VerboseOutput:                         0,
		IgnoreContainerCpuLimitRequirement:    false,
		IgnoreContainerMemoryLimitRequirement: false,
		IgnoredTests:                          nil,
		EnabledOptionalTests:                  nil,
		UseIgnoreChecksAnnotation:             false,
		UseOptionalChecksAnnotation:           false,
	}

	p, err := parser.New()
	if err != nil {
		log.Error().Err(err).Msg("failed to create parser")
		return nil, err
	}
	parsedFiles, err := p.ParseFiles(cnf)
	if err != nil {
		log.Error().Err(err).Msg("failed to parse files")
		return nil, err
	}

	scoreCard, err := score.Score(parsedFiles, cnf)
	if err != nil {
		return nil, err
	}
	return scoreCard, nil
}

func AddKubeScoreAttributesDeployment(deployment *appsv1.Deployment) map[string][]string {
	apiVersion := "apps/v1"
	kind := "Deployment"
	if deployment.APIVersion != "" {
		apiVersion = deployment.APIVersion
	}
	if deployment.Kind != "" {
		kind = deployment.Kind
	}
	return AddKubeScoreAttributes(deployment, deployment.Namespace, deployment.Name, apiVersion, kind)
}

func AddKubeScoreAttributesDaemonSet(daemonSet *appsv1.DaemonSet) map[string][]string {
	apiVersion := "apps/v1"
	kind := "Deployment"
	if daemonSet.APIVersion != "" {
		apiVersion = daemonSet.APIVersion
	}
	if daemonSet.Kind != "" {
		kind = daemonSet.Kind
	}
	return AddKubeScoreAttributes(daemonSet, daemonSet.Namespace, daemonSet.Name, apiVersion, kind)
}
func AddKubeScoreAttributesStatefulSet(statefulSet *appsv1.StatefulSet) map[string][]string {
	apiVersion := "apps/v1"
	kind := "Deployment"
	if statefulSet.APIVersion != "" {
		apiVersion = statefulSet.APIVersion
	}
	if statefulSet.Kind != "" {
		kind = statefulSet.Kind
	}
	return AddKubeScoreAttributes(statefulSet, statefulSet.Namespace, statefulSet.Name, apiVersion, kind)
}
func AddKubeScoreAttributes(obj runtime.Object, namespace string, name string, apiVersion string, kind string) map[string][]string {
	attributes := make(map[string][]string)
	manifestBuf := new(bytes.Buffer)
	err := serializer.Encode(obj, manifestBuf)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to marshal obj %s/%s", namespace, name)
	} else {

		manifest := "apiVersion: " + apiVersion + "\nkind: " + kind + "\n" + manifestBuf.String()
		scoreCard, err := getKubeScore(manifest)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to get kube-score for obj %s/%s", namespace, name)
		} else {
			for _, scoredObject := range *scoreCard {
				for _, check := range scoredObject.Checks {
					attributes["k8s.kube-score."+check.Check.ID+".grade"] = []string{gradeToString(check)}
					attributes["k8s.kube-score."+check.Check.ID+".comment"] = []string{check.Check.Comment}
				}
			}
		}
	}
	return attributes
}

func gradeToString(check scorecard.TestScore) string {
	switch check.Grade {
	case scorecard.GradeCritical:
		return "CRITICAL"
	case scorecard.GradeWarning:
		return "WARNING"
	case scorecard.GradeAlmostOK, scorecard.GradeAllOK:
		return "OK"
	default:
		if check.Skipped {
			return "SKIPPED"
		}
		return "UNKNOWN"
	}
}

type inputReader struct {
	*strings.Reader
}

func (inputReader) Name() string {
	return "input"
}
