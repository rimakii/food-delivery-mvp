package domain

import "testing"

func TestOrderStatus_Valid(t *testing.T) {
	tests := []struct {
		name   string
		status OrderStatus
		want   bool
	}{
		{"new", OrderStatusNew, true},
		{"confirmed", OrderStatusConfirmed, true},
		{"cancelled", OrderStatusCancelled, true},
		{"unknown status", OrderStatus("delivered"), false},
		{"empty status", OrderStatus(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.Valid(); got != tt.want {
				t.Errorf("Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}
