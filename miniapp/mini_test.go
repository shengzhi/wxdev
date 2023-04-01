package miniapp

import (
	"testing"
)

func TestGetPhone(t *testing.T) {
	cli := NewClient("wxe419ff18d9104259", "5bfe5a4cf3c873ea71cd64fe912c3d46", WithTokenServer("http://10.1.3.153:51999"))
	phone, err := cli.GetPhoneNumber("994721ef0f19dddeb54251d6f7e9ae4fca8158f927c36ee984a6f15380e4664c")
	if err != nil {
		t.Logf("error:%s", err.Error())
		t.Fail()
	}
	t.Logf("phone:%s", phone.Phone)
}
