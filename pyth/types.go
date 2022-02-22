//  Copyright 2022 Blockdaemon Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pyth

import (
	"errors"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
)

// Magic is the 32-bit number prefixed on each account.
const Magic = uint32(0xa1b2c3d4)

// V2 identifies the version 2 data format stored in an account.
const V2 = uint32(2)

// The Account type enum identifies what each Pyth account stores.
const (
	AccountTypeUnknown = uint32(iota)
	AccountTypeMapping
	AccountTypeProduct
	AccountTypePrice
)

type AccountHeader struct {
	Magic       uint32
	Version     uint32
	AccountType uint32
}

func (h AccountHeader) Valid() bool {
	return h.Magic == Magic && h.Version == V2
}

// PeekAccount determines the account type given the account's data bytes.
func PeekAccount(data []byte) uint32 {
	decoder := bin.NewBinDecoder(data)
	var header AccountHeader
	if decoder.Decode(&header) != nil || !header.Valid() {
		return AccountTypeUnknown
	}
	return header.AccountType
}

type Ema struct {
	Val   int64
	Numer int64
	Denom int64
}

type PriceInfo struct {
	Price   int64
	Conf    uint64
	Status  uint32
	CorpAct uint32
	PubSlot uint64
}

type PriceComp struct {
	Publisher solana.PublicKey
	Agg       PriceInfo
	Latest    PriceInfo
}

type PriceAccount struct {
	AccountHeader
	Size       uint32
	PriceType  uint32
	Exponent   int32
	Num        uint32
	NumQt      uint32
	LastSlot   uint64
	ValidSlot  uint64
	Twap       Ema
	Twac       Ema
	Drv1, Drv2 int64
	Product    solana.PublicKey
	Next       solana.PublicKey
	PrevSlot   uint64
	PrevPrice  int64
	PrevConf   uint64
	Drv3       int64
	Agg        PriceInfo
	Components [32]PriceComp
}

// UnmarshalBinary decodes the price account from the on-chain format.
func (p *PriceAccount) UnmarshalBinary(buf []byte) error {
	decoder := bin.NewBinDecoder(buf)
	if err := decoder.Decode(p); err != nil {
		return err
	}
	if !p.AccountHeader.Valid() {
		return errors.New("invalid account")
	}
	if p.AccountType != AccountTypePrice {
		return errors.New("not a price account")
	}
	return nil
}

// GetComponent returns the first price component with the given publisher key. Might return nil.
func (p *PriceAccount) GetComponent(publisher *solana.PublicKey) *PriceComp {
	for i := range p.Components {
		if p.Components[i].Publisher == *publisher {
			return &p.Components[i]
		}
	}
	return nil
}
