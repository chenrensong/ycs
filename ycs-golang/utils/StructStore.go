package utils

import (
	"container/list"
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/chenrensong/ygo/lib0/encoding"
	"github.com/chenrensong/ygo/structs"
	"github.com/chenrensong/ygo/utils"
)

// StructStore 管理 Yjs 文档中的所有结构体
type StructStore struct {
	mu                   sync.RWMutex
	clients              map[uint64][]structs.AbstractStruct
	pendingStructRefs    map[uint64]*pendingClientStructRef
	pendingStack         *list.List
	pendingDeleteReaders []DSDecoderV2
}

type pendingClientStructRef struct {
	nextReadOp int
	refs       []structs.AbstractStruct
}

// NewStructStore 创建新的 StructStore 实例
func NewStructStore() *StructStore {
	return &StructStore{
		clients:              make(map[uint64][]structs.AbstractStruct),
		pendingStructRefs:    make(map[uint64]*pendingClientStructRef),
		pendingStack:         list.New(),
		pendingDeleteReaders: make([]DSDecoderV2, 0),
	}
}

// GetStateVector 获取状态向量
func (s *StructStore) GetStateVector() map[uint64]uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	state := make(map[uint64]uint64, len(s.clients))
	for client, structs := range s.clients {
		if len(structs) > 0 {
			last := structs[len(structs)-1]
			state[client] = last.ID().Clock + uint64(last.Length())
		}
	}
	return state
}

// GetState 获取指定客户端的状态
func (s *StructStore) GetState(client uint64) uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if structs, exists := s.clients[client]; exists && len(structs) > 0 {
		last := structs[len(structs)-1]
		return last.ID().Clock + uint64(last.Length())
	}
	return 0
}

// IntegrityCheck 检查存储完整性
func (s *StructStore) IntegrityCheck() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for client, structs := range s.clients {
		if len(structs) == 0 {
			return fmt.Errorf("no structs for client %d", client)
		}

		for i := 1; i < len(structs); i++ {
			left := structs[i-1]
			right := structs[i]
			if left.ID().Clock+uint64(left.Length()) != right.ID().Clock {
				return fmt.Errorf("missing struct between %v and %v", left.ID(), right.ID())
			}
		}
	}

	if len(s.pendingDeleteReaders) > 0 || s.pendingStack.Len() > 0 || len(s.pendingStructRefs) > 0 {
		return errors.New("pending items remain")
	}

	return nil
}

// CleanupPendingStructs 清理待处理的结构体引用
func (s *StructStore) CleanupPendingStructs() {
	s.mu.Lock()
	defer s.mu.Unlock()

	var clientsToRemove []uint64

	for client, ref := range s.pendingStructRefs {
		if ref.nextReadOp == len(ref.refs) {
			clientsToRemove = append(clientsToRemove, client)
		} else {
			ref.refs = ref.refs[ref.nextReadOp:]
			ref.nextReadOp = 0
		}
	}

	for _, client := range clientsToRemove {
		delete(s.pendingStructRefs, client)
	}
}

// AddStruct 添加结构体到存储
func (s *StructStore) AddStruct(str structs.AbstractStruct) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	client := str.ID().Client
	structs, exists := s.clients[client]

	if !exists {
		s.clients[client] = []structs.AbstractStruct{str}
		return nil
	}

	last := structs[len(structs)-1]
	if last.ID().Clock+uint64(last.Length()) != str.ID().Clock {
		return errors.New("unexpected struct clock sequence")
	}

	s.clients[client] = append(structs, str)
	return nil
}

// FindIndexSS 在排序的结构体切片中查找索引
func FindIndexSS(structs []structs.AbstractStruct, clock uint64) int {
	if len(structs) == 0 {
		return -1
	}

	left := 0
	right := len(structs) - 1
	mid := structs[right]
	midClock := mid.ID().Clock

	if midClock == clock {
		return right
	}

	// 带枢轴优化的二分查找
	midIndex := int((clock * uint64(right)) / (midClock + uint64(mid.Length()) - 1))
	for left <= right {
		mid = structs[midIndex]
		midClock = mid.ID().Clock

		if midClock <= clock {
			if clock < midClock+uint64(mid.Length()) {
				return midIndex
			}
			left = midIndex + 1
		} else {
			right = midIndex - 1
		}
		midIndex = (left + right) / 2
	}

	return -1
}

// Find 根据ID查找结构体
func (s *StructStore) Find(id utils.ID) (structs.AbstractStruct, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	structs, exists := s.clients[id.Client]
	if !exists {
		return nil, fmt.Errorf("no structs for client %d", id.Client)
	}

	index := FindIndexSS(structs, id.Clock)
	if index < 0 || index >= len(structs) {
		return nil, fmt.Errorf("invalid struct index %d", index)
	}

	return structs[index], nil
}

// FindIndexCleanStart 查找或创建干净的起始索引
func (s *StructStore) FindIndexCleanStart(tx *Transaction, structs []structs.AbstractStruct, clock uint64) (int, error) {
	index := FindIndexSS(structs, clock)
	if index < 0 {
		return -1, errors.New("struct not found")
	}

	str := structs[index]
	if str.ID().Clock < clock {
		if item, ok := str.(*structs.Item); ok {
			splitItem, err := item.SplitItem(tx, int(clock-item.ID().Clock))
			if err != nil {
				return -1, err
			}
			newStructs := make([]structs.AbstractStruct, len(structs)+1)
			copy(newStructs, structs[:index+1])
			newStructs[index+1] = splitItem
			copy(newStructs[index+2:], structs[index+1:])
			return index + 1, nil
		}
	}

	return index, nil
}

// GetItemCleanStart 获取干净的起始项
func (s *StructStore) GetItemCleanStart(tx *Transaction, id utils.ID) (structs.AbstractStruct, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	structs, exists := s.clients[id.Client]
	if !exists {
		return nil, errors.New("client not found")
	}

	index, err := s.FindIndexCleanStart(tx, structs, id.Clock)
	if err != nil {
		return nil, err
	}

	return structs[index], nil
}

// ReplaceStruct 替换现有结构体
func (s *StructStore) ReplaceStruct(old, new structs.AbstractStruct) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	client := old.ID().Client
	structs, exists := s.clients[client]
	if !exists {
		return errors.New("client not found")
	}

	index := FindIndexSS(structs, old.ID().Clock)
	if index < 0 {
		return errors.New("struct not found")
	}

	structs[index] = new
	return nil
}

// ReadAndApplyDeleteSet 读取并应用删除集
func (s *StructStore) ReadAndApplyDeleteSet(decoder encoding.Decoder, tx *Transaction) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	unappliedDs := structs.NewDeleteSet()
	numClients, err := decoder.Reader().ReadVarUint()
	if err != nil {
		return err
	}

	for i := uint64(0); i < numClients; i++ {
		decoder.ResetDsCurVal()

		client, err := decoder.Reader().ReadVarUint()
		if err != nil {
			return err
		}

		numDeletes, err := decoder.Reader().ReadVarUint()
		if err != nil {
			return err
		}

		structs, exists := s.clients[client]
		if !exists {
			structs = make([]structs.AbstractStruct, 0)
		}

		state := s.GetState(client)

		for deleteIndex := uint64(0); deleteIndex < numDeletes; deleteIndex++ {
			clock, err := decoder.ReadDsClock()
			if err != nil {
				return err
			}

			length, err := decoder.ReadDsLength()
			if err != nil {
				return err
			}
			clockEnd := clock + length

			if clock < state {
				if state < clockEnd {
					unappliedDs.Add(client, state, clockEnd-state)
				}

				index := FindIndexSS(structs, clock)
				if index < 0 {
					continue
				}

				str := structs[index]
				if !str.Deleted() && str.ID().Clock < clock {
					item, ok := str.(*structs.Item)
					if !ok {
						continue
					}

					splitItem, err := item.SplitItem(tx, int(clock-item.ID().Clock))
					if err != nil {
						return err
					}

					newStructs := make([]structs.AbstractStruct, len(structs)+1)
					copy(newStructs, structs[:index+1])
					newStructs[index+1] = splitItem
					copy(newStructs[index+2:], structs[index+1:])
					structs = newStructs
					index++
				}

				for index < len(structs) {
					str = structs[index]
					if str.ID().Clock < clockEnd {
						if !str.Deleted() {
							if clockEnd < str.ID().Clock+uint64(str.Length()) {
								item, ok := str.(*structs.Item)
								if !ok {
									index++
									continue
								}

								splitItem, err := item.SplitItem(tx, int(clockEnd-item.ID().Clock))
								if err != nil {
									return err
								}

								newStructs := make([]structs.AbstractStruct, len(structs)+1)
								copy(newStructs, structs[:index+1])
								newStructs[index+1] = splitItem
								copy(newStructs[index+2:], structs[index+1:])
								structs = newStructs
							}

							if err := str.Delete(tx); err != nil {
								return err
							}
						}
						index++
					} else {
						break
					}
				}
			} else {
				unappliedDs.Add(client, clock, clockEnd-clock)
			}
		}
	}

	if unappliedDs.ClientCount() > 0 {
		encoder := encoding.NewDSEncoderV2()
		if err := unappliedDs.Write(encoder); err != nil {
			return err
		}
		s.pendingDeleteReaders = append(s.pendingDeleteReaders, encoding.NewDSDecoderV2(encoder.ToArray()))
	}

	return nil
}

// ResumeStructIntegration 恢复结构体集成
func (s *StructStore) ResumeStructIntegration(tx *Transaction) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.pendingStructRefs) == 0 {
		return nil
	}

	// 按客户端ID排序，先处理高ID
	clientIDs := make([]uint64, 0, len(s.pendingStructRefs))
	for id := range s.pendingStructRefs {
		clientIDs = append(clientIDs, id)
	}
	sort.Slice(clientIDs, func(i, j int) bool {
		return clientIDs[i] > clientIDs[j]
	})

	getNextStructTarget := func() *pendingClientStructRef {
		if len(clientIDs) == 0 {
			return nil
		}

		nextStructsTarget := s.pendingStructRefs[clientIDs[len(clientIDs)-1]]

		for nextStructsTarget.nextReadOp == len(nextStructsTarget.refs) {
			clientIDs = clientIDs[:len(clientIDs)-1]
			if len(clientIDs) > 0 {
				nextStructsTarget = s.pendingStructRefs[clientIDs[len(clientIDs)-1]]
			} else {
				s.pendingStructRefs = make(map[uint64]*pendingClientStructRef)
				return nil
			}
		}

		return nextStructsTarget
	}

	curStructsTarget := getNextStructTarget()
	if curStructsTarget == nil && s.pendingStack.Len() == 0 {
		return nil
	}

	var stackHead structs.AbstractStruct
	if s.pendingStack.Len() > 0 {
		stackHead = s.pendingStack.Remove(s.pendingStack.Back()).(structs.AbstractStruct)
	} else {
		stackHead = curStructsTarget.refs[curStructsTarget.nextReadOp]
		curStructsTarget.nextReadOp++
	}

	state := make(map[uint64]uint64)

	for {
		localClock, exists := state[stackHead.ID().Client]
		if !exists {
			localClock = s.GetState(stackHead.ID().Client)
			state[stackHead.ID().Client] = localClock
		}

		offset := uint64(0)
		if stackHead.ID().Clock < localClock {
			offset = localClock - stackHead.ID().Clock
		}

		if stackHead.ID().Clock+offset != localClock {
			// 缺少前一个消息
			refs, exists := s.pendingStructRefs[stackHead.ID().Client]
			if !exists {
				refs = &pendingClientStructRef{}
			}

			if refs.nextReadOp != len(refs.refs) {
				r := refs.refs[refs.nextReadOp]
				if r.ID().Clock < stackHead.ID().Clock {
					// 将较小时钟的引用放在栈上
					refs.refs[refs.nextReadOp] = stackHead
					s.pendingStack.PushBack(r)

					// 重新排序
					refs.refs = refs.refs[refs.nextReadOp:]
					sort.Slice(refs.refs, func(i, j int) bool {
						return refs.refs[i].ID().Clock < refs.refs[j].ID().Clock
					})
					refs.nextReadOp = 0
					continue
				}
			}

			// 等待缺失的结构体
			s.pendingStack.PushBack(stackHead)
			return nil
		}

		missing := stackHead.GetMissing(tx, s)
		if missing == nil {
			if offset == 0 || offset < uint64(stackHead.Length()) {
				if err := stackHead.Integrate(tx, int(offset)); err != nil {
					return err
				}
				state[stackHead.ID().Client] = stackHead.ID().Clock + uint64(stackHead.Length())
			}

			// 处理下一个栈顶元素
			if s.pendingStack.Len() > 0 {
				stackHead = s.pendingStack.Remove(s.pendingStack.Back()).(structs.AbstractStruct)
			} else if curStructsTarget != nil && curStructsTarget.nextReadOp < len(curStructsTarget.refs) {
				stackHead = curStructsTarget.refs[curStructsTarget.nextReadOp]
				curStructsTarget.nextReadOp++
			} else {
				curStructsTarget = getNextStructTarget()
				if curStructsTarget == nil {
					break
				} else {
					stackHead = curStructsTarget.refs[curStructsTarget.nextReadOp]
					curStructsTarget.nextReadOp++
				}
			}
		} else {
			// 获取包含缺失结构体的读取器
			refs, exists := s.pendingStructRefs[*missing]
			if !exists {
				refs = &pendingClientStructRef{}
			}

			if refs.nextReadOp == len(refs.refs) {
				// 此更新消息因果依赖于另一个更新消息
				s.pendingStack.PushBack(stackHead)
				return nil
			}

			s.pendingStack.PushBack(stackHead)
			stackHead = refs.refs[refs.nextReadOp]
			refs.nextReadOp++
		}
	}

	s.pendingStructRefs = make(map[uint64]*pendingClientStructRef)
	return nil
}

// TryResumePendingDeleteReaders 尝试恢复待处理的删除读取器
func (s *StructStore) TryResumePendingDeleteReaders(tx *Transaction) error {
	s.mu.Lock()
	pendingReaders := make([]encoding.Decoder, len(s.pendingDeleteReaders))
	copy(pendingReaders, s.pendingDeleteReaders)
	s.pendingDeleteReaders = s.pendingDeleteReaders[:0]
	s.mu.Unlock()

	for _, reader := range pendingReaders {
		if err := s.ReadAndApplyDeleteSet(reader, tx); err != nil {
			return err
		}
	}

	return nil
}

// MergeReadStructsIntoPendingReads 合并读取的结构体到待处理读取
func (s *StructStore) MergeReadStructsIntoPendingReads(clientStructsRefs map[uint64][]structs.AbstractStruct) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for client, structRefs := range clientStructsRefs {
		if ref, exists := s.pendingStructRefs[client]; !exists {
			s.pendingStructRefs[client] = &pendingClientStructRef{
				refs: structRefs,
			}
		} else {
			if ref.nextReadOp > 0 {
				ref.refs = ref.refs[ref.nextReadOp:]
				ref.nextReadOp = 0
			}

			ref.refs = append(ref.refs, structRefs...)
			sort.Slice(ref.refs, func(i, j int) bool {
				return ref.refs[i].ID().Clock < ref.refs[j].ID().Clock
			})
		}
	}
}
