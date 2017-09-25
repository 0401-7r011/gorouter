package round_tripper

import (
	"net/http"

	router_http "code.cloudfoundry.org/gorouter/common/http"
	"code.cloudfoundry.org/gorouter/metrics"
	"code.cloudfoundry.org/gorouter/proxy/error_classifiers"
	"code.cloudfoundry.org/gorouter/proxy/utils"
)

type ErrorSpec struct {
	Classifier error_classifiers.Classifier
	Message    string
	Code       int
}

var DefaultErrorSpecs = []ErrorSpec{
	{error_classifiers.AttemptedTLSWithNonTLSBackend, SSLHandshakeMessage, 525},
	{error_classifiers.HostnameMismatch, HostnameErrorMessage, http.StatusServiceUnavailable},
	{error_classifiers.UntrustedCert, InvalidCertificateMessage, 526},
	{error_classifiers.RemoteFailedCertCheck, SSLCertRequiredMessage, 496},
}

type ErrorHandler struct {
	MetricReporter metrics.CombinedReporter
	ErrorSpecs     []ErrorSpec
}

func (eh *ErrorHandler) HandleError(responseWriter utils.ProxyResponseWriter, err error) {
	responseWriter.Header().Set(router_http.CfRouterError, "endpoint_failure")

	eh.writeErrorCode(err, responseWriter)
	responseWriter.Header().Del("Connection")
	responseWriter.Done()
}

func (eh *ErrorHandler) writeErrorCode(err error, responseWriter http.ResponseWriter) {
	for _, spec := range eh.ErrorSpecs {
		if spec.Classifier.Classify(err) {
			http.Error(responseWriter, spec.Message, spec.Code)
			return
		}
	}

	// default case
	http.Error(responseWriter, BadGatewayMessage, http.StatusBadGateway)
	eh.MetricReporter.CaptureBadGateway()
}
