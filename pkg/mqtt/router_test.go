package mqtt

import "testing"

func TestGetControllerListenTopic(t *testing.T) {
	type args struct {
		groupId string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"t", args{"test"}, "$share/test/" + serverSubscribeTopic},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ControllerListenTopic(tt.args.groupId); got != tt.want {
				t.Errorf("ControllerListenTopic() = %v, want %v", got, tt.want)
			}
		})
	}
}
