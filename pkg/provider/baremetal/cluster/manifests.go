package cluster

const (
	tokenFileTemplate = `%s,admin,admin,system:masters
`

	schedulerPolicyConfig = `
{
   "apiVersion" : "v1",
   "extenders" : [
      {
         "apiVersion" : "v1beta1",
         "enableHttps" : false,
         "filterVerb" : "predicates",
         "managedResources" : [
            {
               "ignoredByScheduler" : false,
               "name" : "tencent.com/vcuda-core"
            }
         ],
         "nodeCacheCapable" : false,
         "urlPrefix" : "http://{{.GPUQuotaAdmissionHost}}:3456/scheduler"
      },
      {
         "apiVersion" : "v1beta1",
         "enableHttps" : false,
         "filterVerb" : "filter",
         "BindVerb": "bind",
         "weight": 1,
         "enableHttps": false,
         "managedResources" : [
            {
               "ignoredByScheduler" : true,
               "name" : "tke.cloud.tencent.com/eni-ip"
            }
         ],
         "nodeCacheCapable" : false,
         "urlPrefix" : "http://galaxy-ipam:9040/v1"
      }
   ],
   "kind" : "Policy"
}
`

	auditWebhookConfig = `
apiVersion: v1
kind: Config
clusters:
  - name: tke
    cluster:
      server: {{.AuditBackendAddress}}/apis/audit.tkestack.io/v1/events/sink/{{.ClusterName}}
      insecure-skip-tls-verify: true
current-context: tke
contexts:
  - context:
      cluster: tke
    name: tke
`
)
