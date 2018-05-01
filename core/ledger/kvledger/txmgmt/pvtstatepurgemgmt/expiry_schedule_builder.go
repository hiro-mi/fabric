/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package pvtstatepurgemgmt

import (
	math "math"

	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/privacyenabledstate"
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/statedb"
	"github.com/hyperledger/fabric/core/ledger/pvtdatapolicy"
	"github.com/hyperledger/fabric/core/ledger/util"
)

type expiryScheduleBuilder struct {
	btlPolicy       pvtdatapolicy.BTLPolicy
	scheduleEntries map[expiryInfoKey]*PvtdataKeys
}

func newExpiryScheduleBuilder(btlPolicy pvtdatapolicy.BTLPolicy) *expiryScheduleBuilder {
	return &expiryScheduleBuilder{btlPolicy, make(map[expiryInfoKey]*PvtdataKeys)}
}

func (builder *expiryScheduleBuilder) add(ns, coll, key string, keyHash []byte, versionedValue *statedb.VersionedValue) {
	committingBlk := versionedValue.Version.BlockNum
	expiryBlk := builder.btlPolicy.GetExpiringBlock(ns, coll, committingBlk)
	if isDelete(versionedValue) || neverExpires(expiryBlk) {
		return
	}
	expinfoKey := expiryInfoKey{committingBlk: committingBlk, expiryBlk: expiryBlk}
	pvtdataKeys, ok := builder.scheduleEntries[expinfoKey]
	if !ok {
		pvtdataKeys = newPvtdataKeys()
		builder.scheduleEntries[expinfoKey] = pvtdataKeys
	}
	pvtdataKeys.add(ns, coll, key, keyHash)
}

func isDelete(versionedValue *statedb.VersionedValue) bool {
	return versionedValue.Value == nil
}

func neverExpires(expiryBlk uint64) bool {
	return expiryBlk == math.MaxUint64
}

func (builder *expiryScheduleBuilder) getExpiryInfo() []*expiryInfo {
	var listExpinfo []*expiryInfo
	for expinfoKey, pvtdataKeys := range builder.scheduleEntries {
		expinfoKeyCopy := expinfoKey
		listExpinfo = append(listExpinfo, &expiryInfo{expiryInfoKey: &expinfoKeyCopy, pvtdataKeys: pvtdataKeys})
	}
	return listExpinfo
}

func buildExpirySchedule(
	btlPolicy pvtdatapolicy.BTLPolicy,
	pvtUpdates *privacyenabledstate.PvtUpdateBatch,
	hashedUpdates *privacyenabledstate.HashedUpdateBatch) []*expiryInfo {

	hashedUpdateKeys := hashedUpdates.ToCompositeKeyMap()
	expiryScheduleBuilder := newExpiryScheduleBuilder(btlPolicy)

	// Iterate through the private data updates and for each key add into the expiry schedule
	// i.e., when these private data key and it's hashed-keys are going to be expired
	// Note that the 'hashedUpdateKeys'  may be superset of the pvtUpdates. This is because,
	// the peer may not receive all the private data either because the peer is not eligible for certain private data
	// or because we allow proceeding with the missing private data data
	for pvtUpdateKey, vv := range pvtUpdates.ToCompositeKeyMap() {
		keyHash := util.ComputeStringHash(pvtUpdateKey.Key)
		expiryScheduleBuilder.add(pvtUpdateKey.Namespace, pvtUpdateKey.CollectionName, pvtUpdateKey.Key, keyHash, vv)
		delete(hashedUpdateKeys, privacyenabledstate.HashedCompositeKey{
			Namespace: pvtUpdateKey.Namespace, CollectionName: pvtUpdateKey.CollectionName, KeyHash: string(keyHash)})
	}

	// Add entries for the leftover key hashes i.e., the hashes corresponding to which there is not private key is present
	for hashedUpdateKey, vv := range hashedUpdateKeys {
		expiryScheduleBuilder.add(hashedUpdateKey.Namespace, hashedUpdateKey.CollectionName, "", []byte(hashedUpdateKey.KeyHash), vv)
	}
	return expiryScheduleBuilder.getExpiryInfo()
}
