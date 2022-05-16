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

Define the job within the applications deploy/clowdapp.yml:

```
    jobs:

    - name: org-id-populator
      podSpec:
        image: quay.io/cloudservices/tenant-utils:latest
        command:
          - ./org-id-column-populator
          - -C
          - -a
          - account
          - -o
          - org_id
          - -t
          - connections
          - --ean-translator-addr
          - http://${TENANT_TRANSLATOR_HOST}:${TENANT_TRANSLATOR_PORT}
        env:
          - name: TENANT_TRANSLATOR_HOST
            value: ${TENANT_TRANSLATOR_HOST}
          - name: TENANT_TRANSLATOR_PORT
            value: ${TENANT_TRANSLATOR_PORT}
          - name: LOG_FORMAT
            value: ${LOG_FORMAT}
          - name: LOG_BATCH_FREQUENCY
            value: '1'
        resources:
          limits:
            cpu: 300m
            memory: 1Gi
          requests:
            cpu: 50m
            memory: 512Mi
```

Notice the job has configured the org-id-column-populator to read the database connection
details from the clowder configuration.

The job definition will need to be edited to use the correct table name (-t),
account column name (-a) and org-id column name (-o).

This job will not show up in the cronjob section or the jobs section of the OpenShift UI.
Clowder will not display the job anywhere until the job is kicked off using a Clowder Job 
Invocation (CJI).  Here is a CJI that will run the job:

The `TENANT_TRANSLATOR_HOST` and `TENANT_TRANSLATOR_PORT` environment variables
are set for the stage and prod environments through app interface.  Those environment variables are set
for the ephemeral environments through app-interface as well, but it looks
like they are not set correctly when the application is deployed in the ephemeral envs.
As a result, you will need to set those environment variables manually for testing
in the ephemeral environments:

```
- name: TENANT_TRANSLATOR_HOST
  value: 'apicast.3scale-dev.svc.cluster.local'
- name: TENANT_TRANSLATOR_PORT
  value: '8892'
```

It is possible to use app-interface's gabi utility to query the stage and prod environments
to verify the changes, etc.  The documentation for configuring gabi can be found [here](https://gitlab.cee.redhat.com/service/app-interface/-/blob/master/docs/app-sre/sop/gabi-instances-request.md).

```
---
apiVersion: cloud.redhat.com/v1alpha1
kind: ClowdJobInvocation
metadata:
  name: populate-org-id-column
spec:
  appName: cloud-connector
  jobs:
    - org-id-populator
```

To run the job, save the CJI from above to a file (deploy/run_org_id_populator.yaml) and apply it using `oc apply -f deploy/run_org_id_populator.yaml`.

Unfortunately, it is not possible to manually run the job in the prod environment.  You will need to get a member
of the dev-prod team to kick off the job for you.

You can check the status of the job using the `oc get cji` command:

```
(venv) $ oc get cji
NAME                     COMPLETED
populate-org-id-column   true
tester                   true
```
