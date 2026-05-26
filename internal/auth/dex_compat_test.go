package auth

import "testing"

func TestUnwrapDexSub(t *testing.T) {
	cases := []struct {
		name string
		sub  string
		want string
	}{
		{
			name: "dex_password_alice",
			sub:  "CiQwMDAwMDAwMC0wMDAwLTAwMDAtMDAwMC0wMDAwMDAwMDAwMDESBWxvY2Fs",
			want: "00000000-0000-0000-0000-000000000001",
		},
		{
			name: "dex_password_bob",
			sub:  "CiQwMDAwMDAwMC0wMDAwLTAwMDAtMDAwMC0wMDAwMDAwMDAwMDISBWxvY2Fs",
			want: "00000000-0000-0000-0000-000000000002",
		},
		{
			name: "plain_uuid_unchanged",
			sub:  "00000000-0000-0000-0000-000000000001",
			want: "00000000-0000-0000-0000-000000000001",
		},
		{
			name: "plain_alice_unchanged",
			sub:  "alice",
			want: "alice",
		},
		{
			name: "empty",
			sub:  "",
			want: "",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := unwrapDexSub(tc.sub)
			if got != tc.want {
				t.Errorf("unwrapDexSub(%q) = %q, want %q", tc.sub, got, tc.want)
			}
		})
	}
}
