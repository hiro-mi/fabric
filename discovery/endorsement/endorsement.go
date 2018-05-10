/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package endorsement

import (
	"fmt"

	"github.com/hyperledger/fabric/common/chaincode"
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/hyperledger/fabric/common/graph"
	"github.com/hyperledger/fabric/common/policies"
	"github.com/hyperledger/fabric/gossip/api"
	"github.com/hyperledger/fabric/gossip/common"
	discovery2 "github.com/hyperledger/fabric/gossip/discovery"
	"github.com/hyperledger/fabric/protos/discovery"
	"github.com/hyperledger/fabric/protos/msp"
	"github.com/pkg/errors"
)

var (
	logger = flogging.MustGetLogger("discovery/endorsement")
)

type principalEvaluator interface {
	// SatisfiesPrincipal returns whether a given peer identity satisfies a certain principal
	// on a given channel
	SatisfiesPrincipal(channel string, identity []byte, principal *msp.MSPPrincipal) error

	// MSPOfPrincipal returns the MSP ID of the given principal
	MSPOfPrincipal(principal *msp.MSPPrincipal) string
}

type chaincodeMetadataFetcher interface {
	// ChaincodeMetadata returns the metadata of the chaincode as appears in the ledger,
	// or nil if the channel doesn't exist, or the chaincode isn't found in the ledger
	Metadata(channel string, cc string) *chaincode.Metadata
}

type policyFetcher interface {
	// PolicyByChaincode returns a policy that can be inquired which identities
	// satisfy it
	PolicyByChaincode(channel string, cc string) policies.InquireablePolicy
}

type gossipSupport interface {
	// IdentityInfo returns identity information about peers
	IdentityInfo() api.PeerIdentitySet

	// PeersOfChannel returns the NetworkMembers considered alive
	// and also subscribed to the channel given
	PeersOfChannel(common.ChainID) discovery2.Members

	// Peers returns the NetworkMembers considered alive
	Peers() discovery2.Members
}

type endorsementAnalyzer struct {
	gossipSupport
	principalEvaluator
	policyFetcher
	chaincodeMetadataFetcher
}

// NewEndorsementAnalyzer constructs an NewEndorsementAnalyzer out of the given support
func NewEndorsementAnalyzer(gs gossipSupport, pf policyFetcher, pe principalEvaluator, mf chaincodeMetadataFetcher) *endorsementAnalyzer {
	return &endorsementAnalyzer{
		gossipSupport:            gs,
		policyFetcher:            pf,
		principalEvaluator:       pe,
		chaincodeMetadataFetcher: mf,
	}
}

type peerPrincipalEvaluator func(member discovery2.NetworkMember, principal *msp.MSPPrincipal) bool

// PeersForEndorsement returns an EndorsementDescriptor for a given set of peers, channel, and chaincode
func (ea *endorsementAnalyzer) PeersForEndorsement(chainID common.ChainID, interest *discovery.ChaincodeInterest) (*discovery.EndorsementDescriptor, error) {
	chaincode := interest.Chaincodes[0]
	ccMD := ea.Metadata(string(chainID), chaincode.Name)
	if ccMD == nil {
		return nil, errors.Errorf("No metadata was found for chaincode %s in channel %s", chaincode.Name, string(chainID))
	}
	// Filter out peers that don't have the chaincode installed on them
	chanMembership := ea.PeersOfChannel(chainID).Filter(peersWithChaincode(ccMD))
	channelMembersById := chanMembership.ByID()
	// Choose only the alive messages of those that have joined the channel
	aliveMembership := ea.Peers().Intersect(chanMembership)
	membersById := aliveMembership.ByID()
	// Compute a mapping between the PKI-IDs of members to their identities
	identities := ea.IdentityInfo()
	identitiesOfMembers := computeIdentitiesOfMembers(identities, membersById)

	// Retrieve the policy for the chaincode
	pol := ea.PolicyByChaincode(string(chainID), chaincode.Name)
	if pol == nil {
		logger.Debug("Policy for chaincode '", chaincode, "'doesn't exist")
		return nil, errors.New("policy not found")
	}

	// Compute the combinations of principals (principal sets) that satisfy the endorsement policy
	principalsSets := policies.PrincipalSets(pol.SatisfiedBy())

	// Obtain the MSP IDs of the members of the channel that are alive
	mspIDsOfChannelPeers := mspIDsOfMembers(membersById, identities.ByID())
	// Filter out principal sets that contain MSP IDs for peers that don't have the chaincode(s) installed
	principalsSets = principalsSets.ContainingOnly(func(principal *msp.MSPPrincipal) bool {
		mspID := ea.MSPOfPrincipal(principal)
		_, exists := mspIDsOfChannelPeers[mspID]
		return mspID != "" && exists
	})

	// mapPrincipalsToGroups returns a mapping from principals to their corresponding groups.
	// groups are just human readable representations that mask the principals behind them
	principalGroups := mapPrincipalsToGroups(principalsSets)
	// principalsToPeersGraph computes a bipartite graph (V1 U V2 , E)
	// such that V1 is the peers, V2 are the principals,
	// and each e=(peer,principal) is in E if the peer satisfies the principal
	satGraph := principalsToPeersGraph(principalAndPeerData{
		members: aliveMembership,
		pGrps:   principalGroups,
	}, ea.satisfiesPrincipal(string(chainID), identitiesOfMembers))

	layouts := computeLayouts(principalsSets, principalGroups, satGraph)
	if len(layouts) == 0 {
		return nil, errors.New("cannot satisfy any principal combination")
	}

	criteria := &peerMembershipCriteria{
		possibleLayouts: layouts,
		satGraph:        satGraph,
		chanMemberById:  channelMembersById,
		idOfMembers:     identitiesOfMembers,
	}

	return &discovery.EndorsementDescriptor{
		Chaincode:         chaincode.Name,
		Layouts:           layouts,
		EndorsersByGroups: endorsersByGroup(criteria),
	}, nil
}

func (ea *endorsementAnalyzer) satisfiesPrincipal(channel string, identitiesOfMembers memberIdentities) peerPrincipalEvaluator {
	return func(member discovery2.NetworkMember, principal *msp.MSPPrincipal) bool {
		err := ea.SatisfiesPrincipal(channel, identitiesOfMembers.identityByPKIID(member.PKIid), principal)
		if err == nil {
			// TODO: log the principals in a human readable form
			logger.Debug(member, "satisfies principal", principal)
			return true
		}
		logger.Debug(member, "doesn't satisfy principal", principal, ":", err)
		return false
	}
}

type peerMembershipCriteria struct {
	satGraph        *principalPeerGraph
	idOfMembers     memberIdentities
	chanMemberById  map[string]discovery2.NetworkMember
	possibleLayouts layouts
}

// endorsersByGroup computes a map from groups to peers.
// Each group included, is found in some layout, which means
// that there is some principal combination that includes the corresponding
// group.
// This means that if a group isn't included in the result, there is no
// principal combination (that includes the principal corresponding to the group),
// such that there are enough peers to satisfy the principal combination.
func endorsersByGroup(criteria *peerMembershipCriteria) map[string]*discovery.Peers {
	satGraph := criteria.satGraph
	idOfMembers := criteria.idOfMembers
	chanMemberById := criteria.chanMemberById
	includedGroups := criteria.possibleLayouts.groupsSet()

	res := make(map[string]*discovery.Peers)
	// Map endorsers to their corresponding groups.
	// Iterate the principals, and put the peers into each group that corresponds with a principal vertex
	for grp, principalVertex := range satGraph.principalVertices {
		if _, exists := includedGroups[grp]; !exists {
			// If the current group is not found in any layout, skip the corresponding principal
			continue
		}
		peerList := &discovery.Peers{}
		res[grp] = peerList
		for _, peerVertex := range principalVertex.Neighbors() {
			member := peerVertex.Data.(discovery2.NetworkMember)
			peerList.Peers = append(peerList.Peers, &discovery.Peer{
				Identity:       idOfMembers.identityByPKIID(member.PKIid),
				StateInfo:      chanMemberById[string(member.PKIid)].Envelope,
				MembershipInfo: member.Envelope,
			})
		}
	}
	return res
}

// computeLayouts computes all possible principal combinations
// that can be used to satisfy the endorsement policy, given a graph
// of available peers that maps each peer to a principal it satisfies.
// Each such a combination is called a layout, because it maps
// a group (alias for a principal) to a threshold of peers that need to endorse,
// and that satisfy the corresponding principal.
func computeLayouts(principalsSets []policies.PrincipalSet, principalGroups principalGroupMapper, satGraph *principalPeerGraph) []*discovery.Layout {
	var layouts []*discovery.Layout
	// principalsSets is a collection of combinations of principals,
	// such that each combination (given enough peers) satisfies the endorsement policy.
	for _, principalSet := range principalsSets {
		layout := &discovery.Layout{
			QuantitiesByGroup: make(map[string]uint32),
		}
		// Since principalsSet has repetitions, we first
		// compute a mapping from the principal to repetitions in the set.
		for principal, plurality := range principalSet.UniqueSet() {
			key := principalKey{
				cls:       int32(principal.PrincipalClassification),
				principal: string(principal.Principal),
			}
			// We map the principal to a group, which is an alias for the principal.
			layout.QuantitiesByGroup[principalGroups.group(key)] = uint32(plurality)
		}
		// Check that the layout can be satisfied with the current known peers
		// This is done by iterating the current layout, and ensuring that
		// each principal vertex is connected to at least <plurality> peer vertices.
		if isLayoutSatisfied(layout.QuantitiesByGroup, satGraph) {
			// If so, then add the layout to the layouts, since we have enough peers to satisfy the
			// principal combination
			layouts = append(layouts, layout)
		}
	}
	return layouts
}

func isLayoutSatisfied(layout map[string]uint32, satGraph *principalPeerGraph) bool {
	for grp, plurality := range layout {
		// Do we have more than <plurality> peers connected to the principal?
		if len(satGraph.principalVertices[grp].Neighbors()) < int(plurality) {
			return false
		}
	}
	return true
}

type principalPeerGraph struct {
	peerVertices      []*graph.Vertex
	principalVertices map[string]*graph.Vertex
}

type principalAndPeerData struct {
	members discovery2.Members
	pGrps   principalGroupMapper
}

func principalsToPeersGraph(data principalAndPeerData, satisfiesPrincipal peerPrincipalEvaluator) *principalPeerGraph {
	// Create the peer vertices
	peerVertices := make([]*graph.Vertex, len(data.members))
	for i, member := range data.members {
		peerVertices[i] = graph.NewVertex(string(member.PKIid), member)
	}

	// Create the principal vertices
	principalVertices := make(map[string]*graph.Vertex)
	for pKey, grp := range data.pGrps {
		principalVertices[grp] = graph.NewVertex(grp, pKey.toPrincipal())
	}

	// Connect principals and peers
	for _, principalVertex := range principalVertices {
		for _, peerVertex := range peerVertices {
			// If the current peer satisfies the principal, connect their corresponding vertices with an edge
			principal := principalVertex.Data.(*msp.MSPPrincipal)
			member := peerVertex.Data.(discovery2.NetworkMember)
			if satisfiesPrincipal(member, principal) {
				peerVertex.AddNeighbor(principalVertex)
			}
		}
	}
	return &principalPeerGraph{
		peerVertices:      peerVertices,
		principalVertices: principalVertices,
	}
}

func mapPrincipalsToGroups(principalsSets []policies.PrincipalSet) principalGroupMapper {
	groupMapper := make(principalGroupMapper)
	totalPrincipals := make(map[principalKey]struct{})
	for _, principalSet := range principalsSets {
		for _, principal := range principalSet {
			totalPrincipals[principalKey{
				principal: string(principal.Principal),
				cls:       int32(principal.PrincipalClassification),
			}] = struct{}{}
		}
	}
	for principal := range totalPrincipals {
		groupMapper.group(principal)
	}
	return groupMapper
}

type memberIdentities map[string]api.PeerIdentityType

func (m memberIdentities) identityByPKIID(id common.PKIidType) api.PeerIdentityType {
	return m[string(id)]
}

func computeIdentitiesOfMembers(identitySet api.PeerIdentitySet, members map[string]discovery2.NetworkMember) memberIdentities {
	identitiesByPKIID := make(map[string]api.PeerIdentityType)
	identitiesOfMembers := make(map[string]api.PeerIdentityType, len(members))
	for _, identity := range identitySet {
		identitiesByPKIID[string(identity.PKIId)] = identity.Identity
	}
	for _, member := range members {
		if identity, exists := identitiesByPKIID[string(member.PKIid)]; exists {
			identitiesOfMembers[string(member.PKIid)] = identity
		}
	}
	return identitiesOfMembers
}

// principalGroupMapper maps principals to names of groups
type principalGroupMapper map[principalKey]string

func (mapper principalGroupMapper) group(principal principalKey) string {
	if grp, exists := mapper[principal]; exists {
		return grp
	}
	grp := fmt.Sprintf("G%d", len(mapper))
	mapper[principal] = grp
	return grp
}

type principalKey struct {
	cls       int32
	principal string
}

func (pk principalKey) toPrincipal() *msp.MSPPrincipal {
	return &msp.MSPPrincipal{
		PrincipalClassification: msp.MSPPrincipal_Classification(pk.cls),
		Principal:               []byte(pk.principal),
	}
}

// layouts is an aggregation of several layouts
type layouts []*discovery.Layout

// groupsSet returns a set of groups that the layouts contain
func (l layouts) groupsSet() map[string]struct{} {
	m := make(map[string]struct{})
	for _, layout := range l {
		for grp := range layout.QuantitiesByGroup {
			m[grp] = struct{}{}
		}
	}
	return m
}

func peersWithChaincode(ccMD *chaincode.Metadata) func(member discovery2.NetworkMember) bool {
	return func(member discovery2.NetworkMember) bool {
		if member.Properties == nil {
			return false
		}
		for _, cc := range member.Properties.Chaincodes {
			if cc.Name == ccMD.Name && cc.Version == ccMD.Version {
				return true
			}
		}
		return false
	}
}

func mspIDsOfMembers(membersById map[string]discovery2.NetworkMember, identitiesByID map[string]api.PeerIdentityInfo) map[string]struct{} {
	res := make(map[string]struct{})
	for pkiID := range membersById {
		if identity, exists := identitiesByID[string(pkiID)]; exists {
			res[string(identity.Organization)] = struct{}{}
		}
	}
	return res
}
