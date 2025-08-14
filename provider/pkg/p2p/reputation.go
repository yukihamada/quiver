package p2p

import (
    "math"
    "sync"
    "time"

    "github.com/libp2p/go-libp2p/core/peer"
)

// ReputationScore はノードの評判スコア
type ReputationScore struct {
    Score           float64   `json:"score"`           // 0.0-1.0
    SuccessCount    int64     `json:"success_count"`   // 成功した推論の数
    FailureCount    int64     `json:"failure_count"`   // 失敗した推論の数
    ResponseTime    float64   `json:"response_time"`   // 平均応答時間（秒）
    LastUpdated     time.Time `json:"last_updated"`    // 最終更新時刻
    TotalEarnings   float64   `json:"total_earnings"`  // 総収益
}

// ReputationManager はノードの評判を管理
type ReputationManager struct {
    mu         sync.RWMutex
    scores     map[peer.ID]*ReputationScore
    selfScore  *ReputationScore
}

// NewReputationManager creates a new reputation manager
func NewReputationManager() *ReputationManager {
    return &ReputationManager{
        scores: make(map[peer.ID]*ReputationScore),
        selfScore: &ReputationScore{
            Score:        1.0, // 初期スコアは最高
            LastUpdated:  time.Now(),
        },
    }
}

// UpdateSuccess は成功した推論を記録
func (rm *ReputationManager) UpdateSuccess(peerID peer.ID, responseTime float64, earnings float64) {
    rm.mu.Lock()
    defer rm.mu.Unlock()

    score, exists := rm.scores[peerID]
    if !exists {
        score = &ReputationScore{
            Score: 0.5, // 新規ノードは中間スコアから開始
        }
        rm.scores[peerID] = score
    }

    score.SuccessCount++
    score.TotalEarnings += earnings
    
    // 移動平均で応答時間を更新
    if score.ResponseTime == 0 {
        score.ResponseTime = responseTime
    } else {
        score.ResponseTime = (score.ResponseTime * 0.9) + (responseTime * 0.1)
    }
    
    score.LastUpdated = time.Now()
    score.Score = rm.calculateScore(score)
}

// UpdateFailure は失敗した推論を記録
func (rm *ReputationManager) UpdateFailure(peerID peer.ID) {
    rm.mu.Lock()
    defer rm.mu.Unlock()

    score, exists := rm.scores[peerID]
    if !exists {
        score = &ReputationScore{
            Score: 0.5,
        }
        rm.scores[peerID] = score
    }

    score.FailureCount++
    score.LastUpdated = time.Now()
    score.Score = rm.calculateScore(score)
}

// calculateScore はスコアを計算
func (rm *ReputationManager) calculateScore(score *ReputationScore) float64 {
    if score.SuccessCount+score.FailureCount == 0 {
        return 0.5 // デフォルトスコア
    }

    // 成功率を計算
    successRate := float64(score.SuccessCount) / float64(score.SuccessCount+score.FailureCount)
    
    // 応答時間によるペナルティ（10秒以上は減点）
    timePenalty := 0.0
    if score.ResponseTime > 10.0 {
        timePenalty = math.Min((score.ResponseTime-10.0)/50.0, 0.3)
    }
    
    // 最近のアクティビティボーナス
    activityBonus := 0.0
    daysSinceUpdate := time.Since(score.LastUpdated).Hours() / 24
    if daysSinceUpdate < 1 {
        activityBonus = 0.1
    } else if daysSinceUpdate < 7 {
        activityBonus = 0.05
    }
    
    // 最終スコア計算
    finalScore := (successRate * 0.7) + ((1.0 - timePenalty) * 0.2) + activityBonus
    
    // 0.0-1.0の範囲に制限
    return math.Max(0.0, math.Min(1.0, finalScore))
}

// GetScore はピアのスコアを取得
func (rm *ReputationManager) GetScore(peerID peer.ID) float64 {
    rm.mu.RLock()
    defer rm.mu.RUnlock()

    score, exists := rm.scores[peerID]
    if !exists {
        return 0.5 // 未知のノードはデフォルトスコア
    }

    return score.Score
}

// GetTopNodes は上位N個のノードを取得
func (rm *ReputationManager) GetTopNodes(n int) []peer.ID {
    rm.mu.RLock()
    defer rm.mu.RUnlock()

    type nodeScore struct {
        id    peer.ID
        score float64
    }

    var nodes []nodeScore
    for id, score := range rm.scores {
        // 最近アクティブなノードのみ
        if time.Since(score.LastUpdated) < 24*time.Hour {
            nodes = append(nodes, nodeScore{id: id, score: score.Score})
        }
    }

    // スコアでソート（簡易実装）
    for i := 0; i < len(nodes); i++ {
        for j := i + 1; j < len(nodes); j++ {
            if nodes[j].score > nodes[i].score {
                nodes[i], nodes[j] = nodes[j], nodes[i]
            }
        }
    }

    // 上位N個を返す
    result := make([]peer.ID, 0, n)
    for i := 0; i < len(nodes) && i < n; i++ {
        result = append(result, nodes[i].id)
    }

    return result
}

// GetSelfScore は自分のスコアを取得
func (rm *ReputationManager) GetSelfScore() *ReputationScore {
    rm.mu.RLock()
    defer rm.mu.RUnlock()
    
    return rm.selfScore
}

// UpdateSelfSuccess は自分の成功を記録
func (rm *ReputationManager) UpdateSelfSuccess(responseTime float64, earnings float64) {
    rm.mu.Lock()
    defer rm.mu.Unlock()

    rm.selfScore.SuccessCount++
    rm.selfScore.TotalEarnings += earnings
    
    if rm.selfScore.ResponseTime == 0 {
        rm.selfScore.ResponseTime = responseTime
    } else {
        rm.selfScore.ResponseTime = (rm.selfScore.ResponseTime * 0.9) + (responseTime * 0.1)
    }
    
    rm.selfScore.LastUpdated = time.Now()
    rm.selfScore.Score = rm.calculateScore(rm.selfScore)
}

// ExportScores はスコアをエクスポート
func (rm *ReputationManager) ExportScores() map[string]*ReputationScore {
    rm.mu.RLock()
    defer rm.mu.RUnlock()

    result := make(map[string]*ReputationScore)
    for id, score := range rm.scores {
        result[id.String()] = score
    }
    
    return result
}

// ImportScores はスコアをインポート
func (rm *ReputationManager) ImportScores(scores map[string]*ReputationScore) error {
    rm.mu.Lock()
    defer rm.mu.Unlock()

    for idStr, score := range scores {
        peerID, err := peer.Decode(idStr)
        if err != nil {
            continue
        }
        rm.scores[peerID] = score
    }

    return nil
}