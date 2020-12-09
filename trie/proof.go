package trie

import (
	"WuyaChain/crypto"
	"bytes"
	"fmt"
)

// GetProof constructs a merkle proof for key. The result contains all encoded nodes
// on the path to the value at key. The value itself is also included in the last
// node and can be retrieved by verifying the proof.
//
// If the trie does not contain a value for key, the returned proof contains all
// nodes of the longest existing prefix of the key (at least the root node), ending
// with the node that proves the absence of the key.
func (t *Trie) GetProof(key []byte) (map[string][]byte, error) {
	// Collect all nodes on the path to key.
	key = keybytesToHex(key)
	nodes := make([]noder, 0)
	tn := t.root
	proof := make(map[string][]byte)

	for len(key) > 0 && tn != nil {
		switch n := tn.(type) {
		case *ExtensionNode:
			if len(key) < len(n.Key) || !bytes.Equal(n.Key, key[:len(n.Key)]) {
				// The trie doesn't contain the key.
				tn = nil
			} else {
				tn = n.NextNode

				// for ExtensionNode, skip the prefix with len(n.Key),
				key = key[len(n.Key):]
			}
			nodes = append(nodes, n)
		case *BranchNode:
			tn = n.Children[key[0]]

			// for BranchNode, just skip one prefix char,
			key = key[1:]
			nodes = append(nodes, n)
		case hashNode:
			var err error
			tn, err = t.loadNode(n)
			if err != nil {
				return proof, fmt.Errorf("unhandled trie error: %s", err)
			}
		case *LeafNode:
			tn = nil
			if len(key) >= len(n.Key) && bytes.Equal(n.Key, key[:len(n.Key)]) {
				nodes = append(nodes, n)
			}
		default:
			panic(fmt.Sprintf("%T: invalid node: %v", tn, tn))
		}
	}

	for _, n := range nodes {
		buf := new(bytes.Buffer)
		var sha = crypto.NewKeccak256()
		sha.Reset()
		hash := nodeHash(n, buf, sha, nil, nil)
		encodeNode(n, buf, sha)

		proof[string(hash)] = buf.Bytes()
	}

	return proof, nil
}