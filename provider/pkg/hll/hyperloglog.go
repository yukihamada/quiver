package hll

import (
    "encoding/binary"
    "hash/fnv"
    "math"
    "sync"
)

const (
    // HLL++ constants
    precision = 14 // 2^14 = 16384 registers
    regSize   = 6  // 6 bits per register (max leading zeros = 64)
    alpha     = 0.7213 / (1 + 1.079/float64(1<<precision))
)

// HyperLogLog represents a HLL++ sketch for cardinality estimation
type HyperLogLog struct {
    mu        sync.RWMutex
    registers []uint8
}

// New creates a new HyperLogLog sketch
func New() *HyperLogLog {
    return &HyperLogLog{
        registers: make([]uint8, 1<<precision),
    }
}

// Add adds an element to the HLL sketch
func (hll *HyperLogLog) Add(data []byte) {
    hash := hash64(data)
    
    // First p bits determine the register
    idx := hash >> (64 - precision)
    
    // Count leading zeros in the remaining bits + 1
    w := hash << precision
    lz := leadingZeros64(w) + 1
    
    hll.mu.Lock()
    if lz > hll.registers[idx] {
        hll.registers[idx] = uint8(lz)
    }
    hll.mu.Unlock()
}

// AddNode adds a node ID to the sketch
func (hll *HyperLogLog) AddNode(nodeID string) {
    hll.Add([]byte(nodeID))
}

// Estimate returns the estimated cardinality
func (hll *HyperLogLog) Estimate() uint64 {
    hll.mu.RLock()
    defer hll.mu.RUnlock()
    
    sum := 0.0
    zeros := 0
    
    for _, val := range hll.registers {
        if val == 0 {
            zeros++
        }
        sum += math.Pow(2, -float64(val))
    }
    
    m := float64(len(hll.registers))
    estimate := alpha * m * m / sum
    
    // Small range correction
    if estimate <= 2.5*m {
        if zeros != 0 {
            estimate = m * math.Log(m/float64(zeros))
        }
    }
    
    // Large range correction
    if estimate > (1.0/30.0)*math.Pow(2, 32) {
        estimate = -math.Pow(2, 32) * math.Log(1-estimate/math.Pow(2, 32))
    }
    
    return uint64(estimate)
}

// Merge combines two HLL sketches
func (hll *HyperLogLog) Merge(other *HyperLogLog) {
    hll.mu.Lock()
    other.mu.RLock()
    defer hll.mu.Unlock()
    defer other.mu.RUnlock()
    
    for i := range hll.registers {
        if other.registers[i] > hll.registers[i] {
            hll.registers[i] = other.registers[i]
        }
    }
}

// Export returns the raw registers for network transmission
func (hll *HyperLogLog) Export() []byte {
    hll.mu.RLock()
    defer hll.mu.RUnlock()
    
    // Pack 6-bit registers efficiently
    packed := make([]byte, (len(hll.registers)*6+7)/8)
    
    bitPos := 0
    for _, reg := range hll.registers {
        bytePos := bitPos / 8
        bitOffset := bitPos % 8
        
        packed[bytePos] |= reg << (2 - bitOffset)
        if bitOffset > 2 {
            packed[bytePos+1] |= reg >> (bitOffset - 2)
        }
        
        bitPos += 6
    }
    
    return packed
}

// Import loads registers from packed data
func (hll *HyperLogLog) Import(data []byte) error {
    hll.mu.Lock()
    defer hll.mu.Unlock()
    
    expectedLen := (len(hll.registers)*6 + 7) / 8
    if len(data) != expectedLen {
        return &HLLError{"invalid data length"}
    }
    
    bitPos := 0
    for i := range hll.registers {
        bytePos := bitPos / 8
        bitOffset := bitPos % 8
        
        reg := data[bytePos] >> (2 - bitOffset) & 0x3F
        if bitOffset > 2 && bytePos+1 < len(data) {
            reg |= data[bytePos+1] << (bitOffset - 2) & 0x3F
        }
        
        hll.registers[i] = reg
        bitPos += 6
    }
    
    return nil
}

// Reset clears all registers
func (hll *HyperLogLog) Reset() {
    hll.mu.Lock()
    defer hll.mu.Unlock()
    
    for i := range hll.registers {
        hll.registers[i] = 0
    }
}

// hash64 computes a 64-bit hash
func hash64(data []byte) uint64 {
    h := fnv.New64a()
    h.Write(data)
    return h.Sum64()
}

// leadingZeros64 counts leading zeros in a 64-bit integer
func leadingZeros64(x uint64) uint8 {
    if x == 0 {
        return 64
    }
    n := uint8(0)
    if x <= 0x00000000FFFFFFFF {
        n += 32
        x <<= 32
    }
    if x <= 0x0000FFFFFFFFFFFF {
        n += 16
        x <<= 16
    }
    if x <= 0x00FFFFFFFFFFFFFF {
        n += 8
        x <<= 8
    }
    if x <= 0x0FFFFFFFFFFFFFFF {
        n += 4
        x <<= 4
    }
    if x <= 0x3FFFFFFFFFFFFFFF {
        n += 2
        x <<= 2
    }
    if x <= 0x7FFFFFFFFFFFFFFF {
        n += 1
    }
    return n
}

// HLLError represents an HLL-specific error
type HLLError struct {
    msg string
}

func (e *HLLError) Error() string {
    return "hll: " + e.msg
}