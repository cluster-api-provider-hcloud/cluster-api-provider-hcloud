{
  _config+:: {
    local this = self,
    namespace: 'kube-system',
    app: 'kube-keepalived-vip',
    name: $._config.app,
    image: 'docker.io/aledbf/kube-keepalived-vip@sha256:7c8bc1a013369667625aa413c133fb0cbbe5694cfce11ef0f06216e235dee71d',
    notify: {
      local this = self,
      image: 'docker.io/simonswine/hcloud-keepalived-notify@sha256:5747960b2899eb9f7787571baeb74f4552465da7cf5c317bed17b7308d0429b0',
      volumePath: '/notify',
      scriptPath: this.volumePath + '/hcloud-keepalived-notify',
    },
    labels: {
      app: this.app,
      name: this.name,
    },

    hcloudTokenRef+: {
    },
    podsCIDRBlock: '192.168.0.0/16',
    hcloudLoadBalancerIPv4s: [],
    backendService: null,

    template: |||
      {{ $iface := .iface }}{{ $netmask := .netmask }}

      global_defs {
        vrrp_version 3
        vrrp_iptables {{ .iptablesChain }}
        #get rid of warning:  default user 'keepalived_script' for script execution does not exist - please create
        script_user root
        enable_script_security

      }


      #Check if the VIP list is empty
      {{ if not .vipIsEmpty }}


      {{ if .proxyMode }}
      vrrp_script chk_haproxy {
        script "/haproxy-check.sh"
        interval 1
      }
      {{ end }}

      vrrp_instance vips {
        state BACKUP
        interface {{ $iface }}
        virtual_router_id {{ .vrid }}
        priority {{ .priority }}
        nopreempt
        advert_int 1

        track_interface {
          {{ $iface }}
        }

        {{ if .notify }} notify {{ .notify }} {{ end }}

        {{ if .useUnicast }}
        unicast_src_ip {{ .myIP }}
        unicast_peer { {{ range .nodes }}
          {{ . }}{{ end }}
        }
        {{ end }}

        virtual_ipaddress { {{ range .vips }}
          {{ . }}{{ end }}
        }

      {{ if .proxyMode }}
        # In proxy mode there is no need to create virtual servers
        track_script {
          chk_haproxy
        }
      {{ end }}

      }

      {{ if not .proxyMode }}
      {{ range $i, $svc := .svcs }}
      {{ if eq $svc.LVSMethod "VIP" }}
      # VIP Service with no pods: {{ $svc.IP }}
      {{ else }}
      # Service: {{ $svc.Name }}
      virtual_server {{ $svc.IP }} {{ $svc.Port }} {
        delay_loop 5
        lvs_sched wlc
        lvs_method {{ $svc.LVSMethod }}
        persistence_timeout 1800
        protocol {{ $svc.Protocol }}

        {{ range $j, $backend := $svc.Backends }}
        real_server {{ $backend.IP }} {{ $backend.Port }} {
          weight 1
          TCP_CHECK {
            connect_port {{ $backend.Port }}
            connect_timeout 3
          }
        }
        {{ end }}
      }
      {{ end }}
      {{ end }}

      #End if vip list is empty
      {{ end }}
      {{ end }}
    |||,

  },

  /* TODO: Apply me later
  */

  metadata:: {
    name: $._config.name,
    namespace: $._config.namespace,
    labels: $._config.labels,
  },

  configMapTemplate: {
    local this = self,
    apiVersion: 'v1',
    kind: 'ConfigMap',
    metadata: $.metadata {
      name: '%(name)s-template-%(hash)s' % {
        name: $.metadata.name,
        hash: std.substr(
          std.md5(std.manifestJsonEx(this.data, '')),
          0,
          8,
        ),
      },
    },
    data: {
      'keepalived.tmpl': $._config.template,
    },
  },

  configMap: {
    apiVersion: 'v1',
    kind: 'ConfigMap',
    metadata: $.metadata,
    data: {
      [x]: $._config.backendService
      for x in $._config.hcloudLoadBalancerIPv4s
    },
  },

  serviceAccount: {
    apiVersion: 'v1',
    kind: 'ServiceAccount',
    metadata: $.metadata,
  },

  clusterRole: {
    apiVersion: 'rbac.authorization.k8s.io/v1',
    kind: 'ClusterRole',
    metadata: std.prune($.metadata {
      namespace: null,
    }),
    rules: [
      {
        apiGroups: [
          '',
        ],
        resources: [
          'pods',
          'nodes',
          'endpoints',
          'services',
          'configmaps',
        ],
        verbs: [
          'get',
          'list',
          'watch',
        ],
      },
    ],
  },

  clusterRoleBinding: {
    apiVersion: 'rbac.authorization.k8s.io/v1',
    kind: 'ClusterRoleBinding',
    metadata: $.metadata,
    roleRef: {
      apiGroup: 'rbac.authorization.k8s.io',
      kind: 'ClusterRole',
      name: $.clusterRole.metadata.name,
    },
    subjects: [
      {
        kind: 'ServiceAccount',
        name: $.serviceAccount.metadata.name,
        namespace: $.serviceAccount.metadata.namespace,
      },
    ],
  },

  notifyContainer:: {
    image: $._config.notify.image,
    name: 'install-notify',
    command: [
      '/bin/sh',
      '-c',
      |||
        #!/bin/sh
        set -eu

        # copy binary into empty dir
        cp -v /hcloud-keepalived-notify %(script)s
        chmod 0700 %(volume)s
        chmod 0755 %(script)s
      ||| % {
        volume: $._config.notify.volumePath,
        script: $._config.notify.scriptPath,
      },
    ],
    volumeMounts: [
      {
        mountPath: $._config.notify.volumePath,
        name: 'notify',
      },
    ],
  },

  daemonSet: {
    apiVersion: 'apps/v1',
    kind: 'DaemonSet',
    metadata: $.metadata,
    spec: {
      selector: {
        matchLabels: $._config.labels,
      },
      template: {
        metadata: {
          labels: $._config.labels,
        },
        spec: {
          hostNetwork: true,
          serviceAccountName: $._config.name,
          initContainers: [
            $.notifyContainer,
          ],
          containers: [
            {
              image: $._config.image,
              name: $._config.app,
              imagePullPolicy: 'IfNotPresent',
              livenessProbe: {
                httpGet: {
                  path: '/health',
                  port: 8080,
                },
                initialDelaySeconds: 15,
                timeoutSeconds: 3,
              },
              securityContext: {
                privileged: true,
              },
              volumeMounts: [
                {
                  mountPath: '/lib/modules',
                  name: 'modules',
                  readOnly: true,
                },
                {
                  mountPath: '/dev',
                  name: 'dev',
                },
                {
                  mountPath: '/notify',
                  name: 'notify',
                },
                {
                  mountPath: '/keepalived.tmpl',
                  name: 'template',
                  subPath: 'keepalived.tmpl',
                },
              ],
              env: [
                {
                  name: 'POD_NAME',
                  valueFrom: {
                    fieldRef: {
                      fieldPath: 'metadata.name',
                    },
                  },
                },
                {
                  name: 'POD_NAMESPACE',
                  valueFrom: {
                    fieldRef: {
                      fieldPath: 'metadata.namespace',
                    },
                  },
                },
                {
                  name: 'NOTIFY_NODE_NAME',
                  valueFrom: {
                    fieldRef: {
                      fieldPath: 'spec.nodeName',
                    },
                  },
                },
                {
                  name: 'KEEPALIVED_NOTIFY',
                  value: $._config.notify.scriptPath,
                },
                // This uses the PID1 stderr
                {
                  name: 'NOTIFY_LOG_PATH',
                  value: '/proc/1/fd/1',
                },
              ] + if std.length($._config.hcloudLoadBalancerIPv4s) == 0 then [] else [
                {
                  name: 'NOTIFY_FLOATING_IPS',
                  value: std.join(',', $._config.hcloudLoadBalancerIPv4s),
                },
              ] + if $._config.hcloudTokenRef == null then [] else [
                {
                  name: 'NOTIFY_HCLOUD_TOKEN',
                } + $._config.hcloudTokenRef,
              ],
              args: [
                '--services-configmap=%s/%s' % [$._config.namespace, $.configMap.metadata.name],
                '--use-unicast=true',
              ],
            },
          ],
          volumes: [
            {
              name: 'modules',
              hostPath: {
                path: '/lib/modules',
              },
            },
            {
              name: 'dev',
              hostPath: {
                path: '/dev',
              },
            },
            {
              name: 'notify',
              emptyDir: {},
            },
            {
              name: 'template',
              configMap: {
                name: $.configMapTemplate.metadata.name,
              },
            },
          ],
        },
      },
    },
  },
}
