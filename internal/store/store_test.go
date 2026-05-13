package store

import (
	"strings"
	"testing"
)

func TestSQLConstants_HavePlaceholders(t *testing.T) {
	cases := map[string]string{
		"SearchCustomer": SQLSearchCustomer,
		"LookupCustomer": SQLLookupCustomer,
		"ListOrders":     SQLListOrders,
		"GetOrder":       SQLGetOrder,
		"GetOrderItems":  SQLGetOrderItems,
		"InsertAudit":    SQLInsertAudit,
	}
	for name, sql := range cases {
		if !strings.Contains(sql, "$1") {
			t.Errorf("%s missing placeholder $1", name)
		}
		if strings.Contains(sql, "%s") || strings.Contains(sql, "%v") {
			t.Errorf("%s contains a format-verb (possible injection vector)", name)
		}
	}
}
