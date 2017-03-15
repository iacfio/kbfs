// Copyright 2016 Keybase Inc. All rights reserved.
// Use of this source code is governed by a BSD
// license that can be found in the LICENSE file.

package libkbfs

import (
	"fmt"

	"github.com/keybase/kbfs/kbfsblock"
	"golang.org/x/net/context"
	"golang.org/x/net/trace"
)

// blockGetter provides the API for the block retrieval worker to obtain blocks.
type blockGetter interface {
	getBlock(context.Context, KeyMetadata, BlockPointer, Block) error
}

// realBlockGetter obtains real blocks using the APIs available in Config.
type realBlockGetter struct {
	config blockOpsConfig
}

// getBlock implements the interface for realBlockGetter.
func (bg *realBlockGetter) getBlock(ctx context.Context, kmd KeyMetadata, blockPtr BlockPointer, block Block) error {
	tr, trOk := trace.FromContext(ctx)

	if trOk {
		tr.LazyPrintf("getBlock start")
	}

	bserv := bg.config.BlockServer()
	buf, blockServerHalf, err := bserv.Get(
		ctx, kmd.TlfID(), blockPtr.ID, blockPtr.Context)
	if err != nil {
		// Temporary code to track down bad block
		// requests. Remove when not needed anymore.
		if _, ok := err.(kbfsblock.BServerErrorBadRequest); ok {
			panic(fmt.Sprintf("Bad BServer request detected: err=%s, blockPtr=%s",
				err, blockPtr))
		}

		return err
	}

	if trOk {
		tr.LazyPrintf("getBlock end")
	}

	err = assembleBlock(
		ctx, bg.config.keyGetter(), bg.config.Codec(), bg.config.cryptoPure(),
		kmd, blockPtr, block, buf, blockServerHalf)

	if trOk {
		tr.LazyPrintf("assembleBlock end")
	}

	return err
}
