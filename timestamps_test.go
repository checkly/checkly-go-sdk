package checkly_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	checkly "github.com/checkly/checkly-go-sdk/v2"
)

// TestTimestampsDecode verifies that the getters of the resources whose
// timestamps decode into time.Time read them from the created_at and
// updated_at keys the API returns. A null or missing timestamp leaves the
// zero time.
func TestTimestampsDecode(t *testing.T) {
	t.Parallel()
	mustParse := func(s string) time.Time {
		v, err := time.Parse(time.RFC3339Nano, s)
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	tests := []struct {
		name        string
		body        string
		get         func(context.Context, checkly.Client) (createdAt, updatedAt time.Time, err error)
		wantCreated time.Time
		wantUpdated time.Time
	}{
		{
			name: "check",
			body: readFixture(t, "GetCheck.json"),
			get: func(ctx context.Context, c checkly.Client) (time.Time, time.Time, error) {
				check, err := c.GetCheck(ctx, "73d29ea2-6540-4bb5-967e-e07fa2c9465e")
				if err != nil {
					return time.Time{}, time.Time{}, err
				}
				v, ok := check.(*checkly.APICheck)
				if !ok {
					return time.Time{}, time.Time{}, fmt.Errorf("want *APICheck, got %T", check)
				}
				return v.CreatedAt, v.UpdatedAt, nil
			},
			wantCreated: mustParse("2019-05-10T11:14:14.573Z"),
		},
		{
			name: "group",
			body: readFixture(t, "UpdateGroup.json"),
			get: func(ctx context.Context, c checkly.Client) (time.Time, time.Time, error) {
				v, err := c.GetGroupV2(ctx, 1)
				if err != nil {
					return time.Time{}, time.Time{}, err
				}
				return v.CreatedAt, v.UpdatedAt, nil
			},
			wantCreated: mustParse("2020-05-26T11:55:49.670Z"),
			wantUpdated: mustParse("2020-05-26T12:03:20.080Z"),
		},
		{
			name: "snippet",
			body: `{"id":1,"name":"snippet1","script":"script1","created_at":"2026-01-02T03:04:05.678Z","updated_at":"2026-02-03T04:05:06.789Z"}`,
			get: func(ctx context.Context, c checkly.Client) (time.Time, time.Time, error) {
				v, err := c.GetSnippet(ctx, 1)
				if err != nil {
					return time.Time{}, time.Time{}, err
				}
				return v.CreatedAt, v.UpdatedAt, nil
			},
			wantCreated: mustParse("2026-01-02T03:04:05.678Z"),
			wantUpdated: mustParse("2026-02-03T04:05:06.789Z"),
		},
		{
			name: "heartbeat monitor with null updated_at",
			body: `{"id":"850cd536-a63b-4d33-9a12-8b125718f4e2","checkType":"HEARTBEAT","name":"heartbeat","created_at":"2026-01-02T03:04:05.678Z","updated_at":null}`,
			get: func(ctx context.Context, c checkly.Client) (time.Time, time.Time, error) {
				v, err := c.GetHeartbeatMonitor(ctx, "850cd536-a63b-4d33-9a12-8b125718f4e2")
				if err != nil {
					return time.Time{}, time.Time{}, err
				}
				return v.CreatedAt, v.UpdatedAt, nil
			},
			wantCreated: mustParse("2026-01-02T03:04:05.678Z"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := newErrorServer(t, errorReply{status: 200, body: tt.body})
			created, updated, err := tt.get(context.Background(), c)
			if err != nil {
				t.Fatal(err)
			}
			if !created.Equal(tt.wantCreated) {
				t.Errorf("want CreatedAt %v, got %v", tt.wantCreated, created)
			}
			if !updated.Equal(tt.wantUpdated) {
				t.Errorf("want UpdatedAt %v, got %v", tt.wantUpdated, updated)
			}
		})
	}
}
