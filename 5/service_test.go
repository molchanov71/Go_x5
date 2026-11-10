package discount

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApply(t *testing.T) {
	tests := []struct {
		price   int
		percent int
		want    int
	}{
		{-1000, 50, 0},
		{100, -1, 0},
		{100, 0, 100},
		{100, 100, 0},
		{100, 101, 0},
		{1000, 10, 900},
	}

	for _, tt := range tests {
		got, err := Apply(tt.price, tt.percent)
		if tt.price < 0 {
			assert.Error(t, err, ErrNegativePrice)
		} else if tt.percent < 0 || tt.percent > 100 {
			assert.Error(t, err, ErrInvalidPercent)
		} else {
			assert.NoError(t, err)
			require.Equal(t, tt.want, got)
		}
	}
}

func TestValidatePercent(t *testing.T) {
	tests := []int{-1, 10, 100, 101}
	for _, tt := range tests {
		err := ValidatePercent(tt)
		if tt < 0 || tt > 100 {
			assert.Error(t, err, ErrInvalidPercent)
		} else {
			assert.NoError(t, err)
		}
	}
}

func TestValidateFinalPrice(t *testing.T) {
	tests := []struct {
		item Item
		want int
	}{
		{item: Item{Price: 1000, DiscountPercent: 10}, want: 900},
		{item: Item{Price: 1500, DiscountPercent: 50}, want: 750},
		{item: Item{Price: 2000, DiscountPercent: 40}, want: 1200},
	}
	items := make([]Item, 0)
	var wantFinal int
	for _, tt := range tests {
		items = append(items, tt.item)
		wantFinal += tt.want
	}
	final, err := FinalPrice(items)
	assert.NoError(t, err)
	assert.Equal(t, final, wantFinal)
	items = items[:0]
	final, err = FinalPrice(items)
	assert.NoError(t, err)
	assert.Equal(t, final, 0)
	items = []Item{{Price: -1, DiscountPercent: 10}}
	final, err = FinalPrice(items)
	assert.Error(t, err, ErrNegativePrice)
	assert.Equal(t, final, 0)
	items = []Item{{Price: 1000, DiscountPercent: 101}}
	final, err = FinalPrice(items)
	assert.Error(t, err, ErrInvalidPercent)
	assert.Equal(t, final, 0)
}
