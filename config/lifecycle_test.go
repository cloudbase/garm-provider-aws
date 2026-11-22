package config

import "testing"

func TestLifecycleConfiguration(t *testing.T) {
	valid := Config{Region: "eu-central-1", SubnetID: "subnet-example", Credentials: Credentials{CredentialType: AWSCredentialTypeRole}}
	for _, behavior := range []string{"", "stop", "terminate"} {
		c := valid
		c.InstanceInitiatedShutdownBehavior = behavior
		if err := c.Validate(); err != nil {
			t.Fatal(err)
		}
	}
	c := valid
	c.InstanceInitiatedShutdownBehavior = "reboot"
	if c.Validate() == nil {
		t.Fatal("invalid shutdown behavior accepted")
	}
	c = valid
	c.Tags = map[string]string{"GARM_POOL_ID": "override"}
	if c.Validate() == nil {
		t.Fatal("reserved tag override accepted")
	}
	c = valid
	c.Tags = map[string]string{"project": "test"}
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
}
