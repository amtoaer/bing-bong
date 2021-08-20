package model

import (
	"testing"

	"github.com/spf13/viper"
)

func TestInit(t *testing.T) {
	tests := []struct {
		arg  string
		want string
	}{{"botType", "qq"}, {"botUser", "123456"}, {"botPass", "123456"}}
	for _, tt := range tests {
		if got := viper.GetString(tt.arg); got != tt.want {
			t.Errorf("got = %v, want %v", got, tt.want)
		}
	}
}
