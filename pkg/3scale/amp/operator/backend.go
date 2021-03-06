package operator

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	oprand "github.com/3scale/3scale-operator/pkg/crypto/rand"
	"k8s.io/apimachinery/pkg/api/errors"

	v1 "k8s.io/api/core/v1"
)

func (o *OperatorBackendOptionsProvider) GetBackendOptions() (*component.BackendOptions, error) {
	optProv := component.BackendOptionsBuilder{}
	optProv.AppLabel(*o.APIManagerSpec.AppLabel)
	optProv.TenantName(*o.APIManagerSpec.TenantName)
	optProv.WildcardDomain(o.APIManagerSpec.WildcardDomain)

	err := o.setSecretBasedOptions(&optProv)
	if err != nil {
		return nil, err
	}

	o.setResourceRequirementsOptions(&optProv)
	o.setReplicas(&optProv)

	res, err := optProv.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to create Backend Options - %s", err)
	}
	return res, nil
}

func (o *OperatorBackendOptionsProvider) setSecretBasedOptions(b *component.BackendOptionsBuilder) error {
	err := o.setBackendInternalApiOptions(b)
	if err != nil {
		return fmt.Errorf("unable to create Backend Options - %s", err)
	}
	err = o.setBackendListenerOptions(b)
	if err != nil {
		return fmt.Errorf("unable to create Backend Options - %s", err)
	}
	err = o.setBackendRedisOptions(b)
	if err != nil {
		return fmt.Errorf("unable to create Backend Options - %s", err)
	}

	return nil
}

func (o *OperatorBackendOptionsProvider) setResourceRequirementsOptions(b *component.BackendOptionsBuilder) {
	if !*o.APIManagerSpec.ResourceRequirementsEnabled {
		b.ListenerResourceRequirements(v1.ResourceRequirements{})
		b.WorkerResourceRequirements(v1.ResourceRequirements{})
		b.CronResourceRequirements(v1.ResourceRequirements{})
	}
}

func (o *OperatorBackendOptionsProvider) setBackendInternalApiOptions(b *component.BackendOptionsBuilder) error {
	defaultSystemBackendUsername := "3scale_api_user"
	defaultSystemBackendPassword := oprand.String(8)

	currSecret, err := getSecret(component.BackendSecretInternalApiSecretName, o.Namespace, o.Client)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	// when secret is not found, it behaves like an empty secret
	secretData := currSecret.Data
	b.SystemBackendUsername(getSecretDataValueOrDefault(secretData, component.BackendSecretInternalApiUsernameFieldName, defaultSystemBackendUsername))
	b.SystemBackendPassword(getSecretDataValueOrDefault(secretData, component.BackendSecretInternalApiPasswordFieldName, defaultSystemBackendPassword))

	return nil
}

func (o *OperatorBackendOptionsProvider) setBackendListenerOptions(b *component.BackendOptionsBuilder) error {
	currSecret, err := getSecret(component.BackendSecretBackendListenerSecretName, o.Namespace, o.Client)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	secretData := currSecret.Data
	b.ListenerServiceEndpoint(getSecretDataValue(secretData, component.BackendSecretBackendListenerServiceEndpointFieldName))
	b.ListenerRouteEndpoint(getSecretDataValue(secretData, component.BackendSecretBackendListenerRouteEndpointFieldName))

	return nil
}

func (o *OperatorBackendOptionsProvider) setBackendRedisOptions(b *component.BackendOptionsBuilder) error {
	currSecret, err := getSecret(component.BackendSecretBackendRedisSecretName, o.Namespace, o.Client)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	secretData := currSecret.Data
	b.RedisStorageURL(getSecretDataValue(secretData, component.BackendSecretBackendRedisStorageURLFieldName))
	b.RedisQueuesURL(getSecretDataValue(secretData, component.BackendSecretBackendRedisQueuesURLFieldName))
	b.RedisStorageSentinelHosts(getSecretDataValue(secretData, component.BackendSecretBackendRedisStorageSentinelHostsFieldName))
	b.RedisStorageSentinelRole(getSecretDataValue(secretData, component.BackendSecretBackendRedisStorageSentinelRoleFieldName))
	b.RedisQueuesSentinelHosts(getSecretDataValue(secretData, component.BackendSecretBackendRedisQueuesSentinelHostsFieldName))
	b.RedisQueuesSentinelRole(getSecretDataValue(secretData, component.BackendSecretBackendRedisQueuesSentinelRoleFieldName))

	return nil
}

func (o *OperatorBackendOptionsProvider) setReplicas(b *component.BackendOptionsBuilder) {
	b.ListenerReplicas(int32(*o.APIManagerSpec.Backend.ListenerSpec.Replicas))
	b.WorkerReplicas(int32(*o.APIManagerSpec.Backend.WorkerSpec.Replicas))
	b.CronReplicas(int32(*o.APIManagerSpec.Backend.CronSpec.Replicas))
}
