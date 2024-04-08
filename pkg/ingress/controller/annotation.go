package controller

import (
	"strconv"

	"github.com/vngcloud/cloud-provider-vngcloud/pkg/consts"
	"github.com/vngcloud/cloud-provider-vngcloud/pkg/utils"
	"github.com/vngcloud/vngcloud-go-sdk/vngcloud/services/loadbalancer/v2/listener"
	"github.com/vngcloud/vngcloud-go-sdk/vngcloud/services/loadbalancer/v2/loadbalancer"
	"github.com/vngcloud/vngcloud-go-sdk/vngcloud/services/loadbalancer/v2/pool"
	nwv1 "k8s.io/api/networking/v1"
	"k8s.io/klog/v2"
)

const (
	DEFAULT_K8S_SERVICE_ANNOTATION_PREFIX = "vks.vngcloud.vn"
)

// Annotations
const (
	ServiceAnnotationSubnetID              = DEFAULT_K8S_SERVICE_ANNOTATION_PREFIX + "/subnet-id"  // both annotation and cloud-config
	ServiceAnnotationNetworkID             = DEFAULT_K8S_SERVICE_ANNOTATION_PREFIX + "/network-id" // both annotation and cloud-config
	ServiceAnnotationOwnedListeners        = DEFAULT_K8S_SERVICE_ANNOTATION_PREFIX + "/owned-listeners"
	ServiceAnnotationCloudLoadBalancerName = DEFAULT_K8S_SERVICE_ANNOTATION_PREFIX + "/cloud-loadbalancer-name" // set via annotation
	ServiceAnnotationLoadBalancerOwner     = DEFAULT_K8S_SERVICE_ANNOTATION_PREFIX + "/load-balancer-owner"

	// Node annotations
	ServiceAnnotationTargetNodeLabels = DEFAULT_K8S_SERVICE_ANNOTATION_PREFIX + "/target-node-labels"

	// LB annotations
	ServiceAnnotationLoadBalancerID   = DEFAULT_K8S_SERVICE_ANNOTATION_PREFIX + "/load-balancer-id"
	ServiceAnnotationLoadBalancerName = DEFAULT_K8S_SERVICE_ANNOTATION_PREFIX + "/load-balancer-name" // only set via the annotation
	ServiceAnnotationPackageID        = DEFAULT_K8S_SERVICE_ANNOTATION_PREFIX + "/package-id"         // both annotation and cloud-config
	ServiceAnnotationSecurityGroups   = DEFAULT_K8S_SERVICE_ANNOTATION_PREFIX + "/security-groups"
	ServiceAnnotationTags             = DEFAULT_K8S_SERVICE_ANNOTATION_PREFIX + "/tags"
	ServiceAnnotationScheme           = DEFAULT_K8S_SERVICE_ANNOTATION_PREFIX + "/scheme"

	// Listener annotations
	ServiceAnnotationIdleTimeoutClient     = DEFAULT_K8S_SERVICE_ANNOTATION_PREFIX + "/idle-timeout-client"     // both annotation and cloud-config
	ServiceAnnotationIdleTimeoutMember     = DEFAULT_K8S_SERVICE_ANNOTATION_PREFIX + "/idle-timeout-member"     // both annotation and cloud-config
	ServiceAnnotationIdleTimeoutConnection = DEFAULT_K8S_SERVICE_ANNOTATION_PREFIX + "/idle-timeout-connection" // both annotation and cloud-config
	ServiceAnnotationInboundCIDRs          = DEFAULT_K8S_SERVICE_ANNOTATION_PREFIX + "/inbound-cidrs"

	// Pool annotations
	ServiceAnnotationPoolAlgorithm       = DEFAULT_K8S_SERVICE_ANNOTATION_PREFIX + "/pool-algorithm" // both annotation and cloud-config
	ServiceAnnotationEnableStickySession = DEFAULT_K8S_SERVICE_ANNOTATION_PREFIX + "/enable-sticky-session"
	ServiceAnnotationEnableTLSEncryption = DEFAULT_K8S_SERVICE_ANNOTATION_PREFIX + "/enable-tls-encryption"
	ServiceAnnotationHealthcheckPort     = DEFAULT_K8S_SERVICE_ANNOTATION_PREFIX + "/healthcheck-port"

	// Pool healthcheck annotations
	ServiceAnnotationHealthcheckProtocol        = DEFAULT_K8S_SERVICE_ANNOTATION_PREFIX + "/healthcheck-protocol"
	ServiceAnnotationHealthcheckIntervalSeconds = DEFAULT_K8S_SERVICE_ANNOTATION_PREFIX + "/healthcheck-interval-seconds"
	ServiceAnnotationHealthcheckTimeoutSeconds  = DEFAULT_K8S_SERVICE_ANNOTATION_PREFIX + "/healthcheck-timeout-seconds"
	ServiceAnnotationHealthyThresholdCount      = DEFAULT_K8S_SERVICE_ANNOTATION_PREFIX + "/healthy-threshold-count"
	ServiceAnnotationUnhealthyThresholdCount    = DEFAULT_K8S_SERVICE_ANNOTATION_PREFIX + "/unhealthy-threshold-count"

	// Pool healthcheck annotations for HTTP
	ServiceAnnotationHealthcheckPath           = DEFAULT_K8S_SERVICE_ANNOTATION_PREFIX + "/healthcheck-path"
	ServiceAnnotationSuccessCodes              = DEFAULT_K8S_SERVICE_ANNOTATION_PREFIX + "/success-codes"
	ServiceAnnotationHealthcheckHttpMethod     = DEFAULT_K8S_SERVICE_ANNOTATION_PREFIX + "/healthcheck-http-method"
	ServiceAnnotationHealthcheckHttpVersion    = DEFAULT_K8S_SERVICE_ANNOTATION_PREFIX + "/healthcheck-http-version"
	ServiceAnnotationHealthcheckHttpDomainName = DEFAULT_K8S_SERVICE_ANNOTATION_PREFIX + "/healthcheck-http-domain-name"
)

func PointerOf[T any](t T) *T {
	return &t
}

func CreateLoadbalancerOptions(ing *nwv1.Ingress) *loadbalancer.CreateOpts {
	opt := &loadbalancer.CreateOpts{
		Name:      "",
		PackageID: consts.DEFAULT_L7_PACKAGE_ID,
		Scheme:    loadbalancer.CreateOptsSchemeOptInternal,
		SubnetID:  "",
		Type:      loadbalancer.CreateOptsTypeOptLayer7,
	}
	if option, ok := ing.Annotations[ServiceAnnotationLoadBalancerName]; ok {
		opt.Name = option
	}
	if option, ok := ing.Annotations[ServiceAnnotationPackageID]; ok {
		opt.PackageID = option
	}
	if option, ok := ing.Annotations[ServiceAnnotationScheme]; ok {
		switch option {
		case "internal":
			opt.Scheme = loadbalancer.CreateOptsSchemeOptInternal
		case "internet-facing":
			opt.Scheme = loadbalancer.CreateOptsSchemeOptInternet
		default:
			klog.Warningf("Invalid annotation \"%s\" value, must be \"internal\" or \"internet-facing\"", ServiceAnnotationScheme)
		}
	}
	return opt
}

func CreateListenerOptions(ing *nwv1.Ingress, isHTTPS bool) *listener.CreateOpts {
	opt := &listener.CreateOpts{
		ListenerName:                consts.DEFAULT_HTTP_LISTENER_NAME,
		ListenerProtocol:            listener.CreateOptsListenerProtocolOptHTTP,
		ListenerProtocolPort:        80,
		CertificateAuthorities:      nil,
		ClientCertificate:           nil,
		DefaultCertificateAuthority: nil,
		DefaultPoolId:               "",
		TimeoutClient:               50,
		TimeoutMember:               50,
		TimeoutConnection:           5,
		AllowedCidrs:                "0.0.0.0/0",
	}
	if isHTTPS {
		opt.ListenerName = consts.DEFAULT_HTTPS_LISTENER_NAME
		opt.ListenerProtocol = listener.CreateOptsListenerProtocolOptHTTPS
		opt.ListenerProtocolPort = 443

	}
	if ing == nil {
		return opt
	}
	if option, ok := ing.Annotations[ServiceAnnotationIdleTimeoutClient]; ok {
		opt.TimeoutClient = utils.ParseIntAnnotation(option, ServiceAnnotationIdleTimeoutClient, opt.TimeoutClient)
	}
	if option, ok := ing.Annotations[ServiceAnnotationIdleTimeoutMember]; ok {
		opt.TimeoutMember = utils.ParseIntAnnotation(option, ServiceAnnotationIdleTimeoutMember, opt.TimeoutMember)
	}
	if option, ok := ing.Annotations[ServiceAnnotationIdleTimeoutConnection]; ok {
		opt.TimeoutConnection = utils.ParseIntAnnotation(option, ServiceAnnotationIdleTimeoutConnection, opt.TimeoutConnection)
	}
	if option, ok := ing.Annotations[ServiceAnnotationInboundCIDRs]; ok {
		opt.AllowedCidrs = option
	}
	return opt
}

func CreatePoolOptions(ing *nwv1.Ingress) *pool.CreateOpts {
	opt := &pool.CreateOpts{
		PoolName:      "",
		PoolProtocol:  pool.CreateOptsProtocolOptHTTP,
		Stickiness:    PointerOf[bool](false),
		TLSEncryption: PointerOf[bool](false),
		HealthMonitor: pool.HealthMonitor{
			HealthyThreshold:    3,
			UnhealthyThreshold:  3,
			Interval:            30,
			Timeout:             5,
			HealthCheckProtocol: pool.CreateOptsHealthCheckProtocolOptTCP,
		},
		Algorithm: pool.CreateOptsAlgorithmOptRoundRobin,
		Members:   []*pool.Member{},
	}
	if ing == nil {
		return opt
	}
	if option, ok := ing.Annotations[ServiceAnnotationHealthcheckProtocol]; ok {
		switch option {
		case string(pool.CreateOptsHealthCheckProtocolOptTCP), string(pool.CreateOptsHealthCheckProtocolOptHTTP):
			opt.HealthMonitor.HealthCheckProtocol = pool.CreateOptsHealthCheckProtocolOpt(option)
			if option == string(pool.CreateOptsHealthCheckProtocolOptHTTP) {
				opt.HealthMonitor = pool.HealthMonitor{
					HealthyThreshold:    3,
					UnhealthyThreshold:  3,
					Interval:            30,
					Timeout:             5,
					HealthCheckProtocol: pool.CreateOptsHealthCheckProtocolOptHTTP,
					HealthCheckMethod:   PointerOf(pool.CreateOptsHealthCheckMethodOptGET),
					HealthCheckPath:     PointerOf("/"),
					SuccessCode:         PointerOf("200"),
					HttpVersion:         PointerOf(pool.CreateOptsHealthCheckHttpVersionOptHttp1),
					DomainName:          PointerOf(""),
				}
				if option, ok := ing.Annotations[ServiceAnnotationHealthcheckHttpMethod]; ok {
					switch option {
					case string(pool.CreateOptsHealthCheckMethodOptGET),
						string(pool.CreateOptsHealthCheckMethodOptPUT),
						string(pool.CreateOptsHealthCheckMethodOptPOST):
						opt.HealthMonitor.HealthCheckMethod = PointerOf(pool.CreateOptsHealthCheckMethodOpt(option))
					default:
						klog.Warningf("Invalid annotation \"%s\" value, must be one of %s, %s, %s", ServiceAnnotationHealthcheckHttpMethod,
							pool.CreateOptsHealthCheckMethodOptGET,
							pool.CreateOptsHealthCheckMethodOptPUT,
							pool.CreateOptsHealthCheckMethodOptPOST)
					}
				}
				if option, ok := ing.Annotations[ServiceAnnotationHealthcheckPath]; ok {
					opt.HealthMonitor.HealthCheckPath = PointerOf(option)
				}
				if option, ok := ing.Annotations[ServiceAnnotationSuccessCodes]; ok {
					opt.HealthMonitor.SuccessCode = PointerOf(option)
				}
				if option, ok := ing.Annotations[ServiceAnnotationHealthcheckHttpVersion]; ok {
					switch option {
					case string(pool.CreateOptsHealthCheckHttpVersionOptHttp1),
						string(pool.CreateOptsHealthCheckHttpVersionOptHttp1Minor1):
						opt.HealthMonitor.HttpVersion = PointerOf(pool.CreateOptsHealthCheckHttpVersionOpt(option))
					default:
						klog.Warningf("Invalid annotation \"%s\" value, must be one of %s, %s", ServiceAnnotationHealthcheckHttpVersion,
							pool.CreateOptsHealthCheckHttpVersionOptHttp1,
							pool.CreateOptsHealthCheckHttpVersionOptHttp1Minor1)
					}
				}
				if option, ok := ing.Annotations[ServiceAnnotationHealthcheckHttpDomainName]; ok {
					opt.HealthMonitor.DomainName = PointerOf(option)
				}
			}
		default:
			klog.Warningf("Invalid annotation \"%s\" value, must be one of %s, %s", ServiceAnnotationHealthcheckProtocol,
				pool.CreateOptsHealthCheckProtocolOptTCP,
				pool.CreateOptsHealthCheckProtocolOptHTTP)
		}
	}
	if option, ok := ing.Annotations[ServiceAnnotationPoolAlgorithm]; ok {
		switch option {
		case string(pool.CreateOptsAlgorithmOptRoundRobin),
			string(pool.CreateOptsAlgorithmOptLeastConn),
			string(pool.CreateOptsAlgorithmOptSourceIP):
			opt.Algorithm = pool.CreateOptsAlgorithmOpt(option)
		default:
			klog.Warningf("Invalid annotation \"%s\" value, must be one of %s, %s, %s", ServiceAnnotationPoolAlgorithm,
				pool.CreateOptsAlgorithmOptRoundRobin,
				pool.CreateOptsAlgorithmOptLeastConn,
				pool.CreateOptsAlgorithmOptSourceIP)
		}
	}
	if option, ok := ing.Annotations[ServiceAnnotationHealthyThresholdCount]; ok {
		opt.HealthMonitor.HealthyThreshold = utils.ParseIntAnnotation(option, ServiceAnnotationHealthyThresholdCount, opt.HealthMonitor.HealthyThreshold)
	}
	if option, ok := ing.Annotations[ServiceAnnotationUnhealthyThresholdCount]; ok {
		opt.HealthMonitor.UnhealthyThreshold = utils.ParseIntAnnotation(option, ServiceAnnotationUnhealthyThresholdCount, opt.HealthMonitor.UnhealthyThreshold)
	}
	if option, ok := ing.Annotations[ServiceAnnotationHealthcheckTimeoutSeconds]; ok {
		opt.HealthMonitor.Timeout = utils.ParseIntAnnotation(option, ServiceAnnotationHealthcheckTimeoutSeconds, opt.HealthMonitor.Timeout)
	}
	if option, ok := ing.Annotations[ServiceAnnotationHealthcheckIntervalSeconds]; ok {
		opt.HealthMonitor.Interval = utils.ParseIntAnnotation(option, ServiceAnnotationHealthcheckIntervalSeconds, opt.HealthMonitor.Interval)
	}
	if option, ok := ing.Annotations[ServiceAnnotationEnableStickySession]; ok {
		switch option {
		case "true", "false":
			boolValue, _ := strconv.ParseBool(option)
			opt.Stickiness = PointerOf(boolValue)
		default:
			klog.Warningf("Invalid annotation \"%s\" value, must be true or false", ServiceAnnotationEnableStickySession)
		}
	}
	if option, ok := ing.Annotations[ServiceAnnotationEnableTLSEncryption]; ok {
		switch option {
		case "true", "false":
			boolValue, _ := strconv.ParseBool(option)
			opt.TLSEncryption = PointerOf(boolValue)
		default:
			klog.Warningf("Invalid annotation \"%s\" value, must be true or false", ServiceAnnotationEnableTLSEncryption)
		}
	}
	return opt
}
