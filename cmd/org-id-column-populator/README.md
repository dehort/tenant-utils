The org-id-column-populator is simply a binary wrapper around the [Org ID column populator library](../../pkg/tenantconv/README.md).

# Building the org-id-column-populator

```
make
```

# Using the org-id-column-populator

The database connection details can be specified on the command line:

```
org-id-column-populator -H localhost -p 5432 -u insights -w insights -n cloud-connector -t connections -a account -o org_id -b 1 --ean-translator-addr "http://translation-svc:8092"
```

Or the database connection details can be pulled from the clowder config file:

```
ACG_CONFIG=clowder.json org-id-column-populator -C -t connections -a account -o org_id -b 1 --ean-translator-addr "http://translation-svc:8092"
```

The table name (-t), account number column name (-a), org_id column name (-o) and batch size (-b) are configurable.

The`TestBatchTranslator` implementation of the BatchTranslator interface can be used by setting the
`--ean-translator-addr` command line option to `test`.
To use a real instance of the BatchTranslator, set the address of the translation service
using the `--ean-translator-addr` command line option.

A test implementation of the metric recorder, which just prints out the metrics, is used by default.
To use a real instance of the metrics recorder, set the address of the prometheus push gateway using
the `--prometheus-push-addr` command line option.


# Clowder Integration Notes

A pre-built image containing the org-id-column-populator will be available for download
from quay.io.  This image can be deployed as an OpenShift job in order to populate an
org-id column.

Define the job within the applications ClowdApp template:

```
apiVersion: v1
kind: Template
objects:

- apiVersion: cloud.redhat.com/v1alpha1
  kind: ClowdApp
  metadata:
    name: my-app

  //
  // actual ClowdApp definition left out for brewity
  //

    jobs:

    - name: org-id-populator
      podSpec:
        image: quay.io/cloudservices/tenant-utils:latest
        command:
          - ./org-id-column-populator
          - -C
          - -a
          - account // TODO: modify if needed
          - -o
          - org_id // TODO: modify if needed
          - -t
          - connections // TODO: modify if needed
          - --ean-translator-addr
          - http://${TENANT_TRANSLATOR_HOST}:${TENANT_TRANSLATOR_PORT}
          - --prometheus-push-addr
          - ${PROMETHEUS_PUSHGATEWAY}
        env:
          - name: LOG_FORMAT
            value: ${POPULATOR_LOG_FORMAT}
          - name: LOG_BATCH_FREQUENCY
            value: '1s'
        resources:
          limits:
            cpu: 300m
            memory: 1Gi
          requests:
            cpu: 50m
            memory: 512Mi

- apiVersion: cloud.redhat.com/v1alpha1
  kind: ClowdJobInvocation
  metadata:
    name: populate-org-id-column-${POPULATOR_RUN_NUMBER}
  spec:
    appName: my-app // TODO: update this to match the actual app's name
    jobs:
      - org-id-populator

parameters:
  //
  // existing parameters left out for brewity
  //
  - name: TENANT_TRANSLATOR_HOST
  required: true
  - name: TENANT_TRANSLATOR_PORT
    value: '8892'
  - name: POPULATOR_LOG_FORMAT
    value: cloudwatch
  - name: POPULATOR_IMAGE
    value: quay.io/cloudservices/tenant-utils
  - name: POPULATOR_IMAGE_TAG
    value: latest
  - name: POPULATOR_RUN_NUMBER # in case the populator needs to be run more than once increment this parameter to get a new job
    value: "1"
  - name: PROMETHEUS_PUSHGATEWAY
    value: "localhost"

```

Notice the job has configured the org-id-column-populator to read the database connection
details from the clowder configuration.

The job definition will need to be edited to use the correct table name (-t),
account column name (-a) and org-id column name (-o).

## Running the job using app-interface

In order for the org-id-column-populator job to run in app-interface managed environments (stage, prod), these steps need to be taken:

1. A network policy allowing connections from the namespace where the org-id-column-populator job will run to the tenant translator service needs to be defined ([stage](https://gitlab.cee.redhat.com/service/app-interface/-/blob/master/data/services/insights/gateway/namespaces/stage-3scale-stage.yml#L25), [prod](https://gitlab.cee.redhat.com/service/app-interface/-/blob/master/data/services/insights/gateway/namespaces/3scale-prod.yml#L19))
1. A network policy allowing connections from the namespace where the org-id-column-populator job will run to the Prometheus Pushgateway needs to be defined ([stage](https://gitlab.cee.redhat.com/service/app-interface/-/blob/master/data/services/insights/prometheus/insights-push-stage.yml#L19), [prod](https://gitlab.cee.redhat.com/service/app-interface/-/blob/master/data/services/insights/prometheus/insights-push-prod.yml#L19))
1. The list of `managedResourceTypes` in the application's SaaS file needs to be updated to contain `ClowdJobInvocation` ([example](https://gitlab.cee.redhat.com/service/app-interface/-/blob/1f2b7038d32124dc150ed348d509db862305552f/data/services/insights/playbook-dispatcher/deploy.yml#L32))

The `TENANT_TRANSLATOR_HOST`, `TENANT_TRANSLATOR_PORT` and `PROMETHEUS_PUSHGATEWAY` parameters
are set for the stage and prod environments through app interface and will be applied to the template automatically.
Those parameters are set for the ephemeral environments through app-interface as well, but it looks
like they are not set correctly when the application is deployed in the ephemeral envs using bonfire.
As a result, you will need to set those environment variables manually for testing
in the ephemeral environments:

```
bonfire deploy my-app -p my-app/TENANT_TRANSLATOR_HOST=apicast.3scale-dev.svc.cluster.local
```

## Verifying the results

It is possible to use app-interface's gabi utility to query the stage and prod environments
to verify the changes, etc.  The documentation for configuring gabi can be found [here](https://gitlab.cee.redhat.com/service/app-interface/-/blob/master/docs/app-sre/sop/gabi-instances-request.md).

You can check the status of the job using the `oc get cji` command:

```
(venv) $ oc get cji
NAME                     COMPLETED
populate-org-id-column   true
tester                   true
```

Metrics produced by org-id-column-populator will be available in Prometheus under the following metric names:

- `org_id_column_populator_rows_updated`
- `org_id_column_populator_unique_accounts`


## Possible errors

* Error testing connection to the database - x509: certificate relies on legacy Common Name field, use SANs instead

  Starting with Go 1.17, Go will no longer use the CommonName field from the certificate as the server's
  hostname if the certificate does not contain a Subject Alternative Name.  As a result, if the server's
  certificate does not include a Subject Alternative Name, then the cert verification will fail.
  The certificate used by the database must be updated to include a Subject Alternative Name.
  The devprod team have the permission to update the certificates.

* Error updating org id column in the database - exec failed pq: canceling statement due to user request

  The updating of the org-id fields in the database took too long.  The `--db-operation-timeout` command line
  option can be used to adjust the database timeout (the default is 10 seconds).

* Error sending HTTP request: Post \"http://apicast.3scale-stage.svc.cluster.local:8892/internal/orgIds\": context deadline exceeded (Client.Timeout exceeded while awaiting headers)"

  The calls to the tenant translation service took too long.

  This can happen if the network policy has not been configured as described [here](#running-the-job-using-app-interface).

  If the network policy has been configured correctly, then there are two possible ways to resolve this issue:

  1.  reduce the batch size by using the `--batch-size` command line option (the default batch-size is 100)
  1.  increase the tenant translator timeout by using the `--ean-translator-timeout` command line option (the default is 20 seconds)
