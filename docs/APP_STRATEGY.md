# QUIVer アプリケーション戦略

## 推奨トポロジー（最小で強い構成）

### 🅰️ QuicPair（エッジ用・ユーザー向け）
- **対象**: エンドユーザー（個人／チーム）
- **形態**: macOSアプリ＋iOSアプリ（同一ブランドの"1製品ファミリー"）
- **役割**:
  - 端末内LLMをローカル優先で実行（Ollama等）
  - Mac↔iPhoneの超低遅延ペア／E2E
  - スコープスイッチ：Local Only / My Devices / Global(QUIVer)
  - QUIVerにオフロードするときは事前見積＋計算レシート添付
- **課金**: ProでFast Start（プリウォーム）／自動量子化／多デバイス／マネージドTURNなど"運用品質"に課金

### 🅱️ QUIVer Node（プロバイダ向け・供給側）
- **対象**: 計算リソース提供者（個人マイナー／企業GPU／コミュニティ）
- **形態**: デスクトップ常駐 or CLI/Daemon（Linux/Windows/macOS）
- **役割**:
  - ジョブ受信・実行・署名付き計算レシート発行
  - 稼働率・電力・温度のポリシー運用
  - ウォレット／決済（従量）
- **課金**: 従量手数料（ネットワーク側）＋EnterpriseはSLA/私設メッシュ

### 🌐 QUIVer Console（Web/PWA）
- **対象**: Provider/Developer/運用
- **役割**: Explorer（可視化）、ダッシュボード（収益/ノード）、APIポータル、Docs/VDP/Status
- **備考**: Webなので"アプリ"にカウントしない。PWA化でスマホ操作もOK。

## なぜ"2アプリ＋Web"が最適か

1. **ユーザー視点のシンプルさ**:
   "私はQuicPairだけ使えばいい"が一発で伝わる（供給や開発はWebへ）。

2. **開発・運用の集中**:
   UI/UXの磨きどころがぶれない。Nodeはヘッドレスでも成立。

3. **ブランドの整理**:
   QuicPair＝プライベート体験／QUIVer＝分散スケールが崩れない。

4. **配布と法務の合理性**:
   App StoreはQuicPairのみ。NodeはGitHub Release/署名配布で柔軟。

## 機能の"置き場所"早見表

| 機能 | QuicPair | QUIVer Node | Console(Web) |
|------|----------|-------------|--------------|
| ローカル推論（Ollama/MLX） | ✅ | ー | ー |
| Mac↔iPhoneペア/E2E | ✅ | ー | ー |
| スコープ切替（Local/My/Global） | ✅ | ー | ー |
| QUIVerオフロード＆見積 | ✅ | ー | ー |
| 計算レシートの表示 | ✅（結果に添付） | ー | ✅（検証/履歴） |
| ノード稼働/温度/電力ポリシー | ー | ✅ | ✅（設定UI） |
| 収益と出金 | ー | ー | ✅ |
| Explorer（地図/ランキング） | ー | ー | ✅ |
| APIキー/SDK配布/Docs | ー | ー | ✅ |
| SSO/MDM/ポリシー配布 | ー | ー | ✅（Team/Ent） |

## プラットフォーム別の優先度

1. **QuicPair**: macOS / iOS（現方針維持）
   - iOSはクライアント中心。必要なら超小型モデルのオンデバイスfallback（例：Gemma 270M級）を後付け。

2. **QUIVer Node**: Linux（最優先） → Windows → macOS
   - まずLinuxでクラスタ/自宅サーバを抑え、次にWindowsのGPU層を開拓。

3. **Console（Web）**: Explorer/ダッシュボード/API/Docs/Status/VDPを順次拡張。

## 3本目の"ネイティブ"を増やす判断基準

以下のトリガーを2つ以上満たしたら、3本目（例：QUIVer Provider DesktopのGUI専用アプリ）を検討:
- Providerの非CLI比率> 40%（CLIがつらいという声が多い）
- サポート起因のオンボード工数が月30時間超
- 家庭用GPU/CPU Providerの日次アクティブが1,000超（UX投資で歩留まり改善余地）
- 収益の>30%が"家庭Provider"から発生（GUI投資の回収目処）

## リリース運用ロードマップ

- **R1（現状〜）**: 2アプリ＋Console最小
  - QuicPairのPro機能（Fast Start/自動量子化/マネージドTURN/多デバイス）
  - NodeはCLI/Daemon＋署名レシート
  - ConsoleはExplorer簡易＋Providerダッシュボードβ

- **R2**: Windows Node／Console本格化（収益/通知/アラート）

- **R3**: QuicPair→QUIVerの1クリック見積/上限設定、レシートUX磨き

- **R4（需要次第）**: Node GUI（3本目） or モバイルProvider（制約強）検証

## 体験原則（これだけは守る）

1. **Private first**: QuicPairは常にLocal Onlyが既定。外に出す時は明示UI＋見積。
2. **Receipt by default**: QUIVer経由の出力は計算レシートを必ず添付。
3. **One‑tap scale**: QuicPairから1タップでスケール（スコープスイッチ）。
4. **No account lock‑in**: QuicPair単体でも一生使える（ローカル機能は無料維持）。

## 最終結論

- **2アプリでOK**: QuicPair（エッジ）とQUIVer Node（供給）に役割を分け、管理・可視化はWebコンソールに寄せる。
- 3本目のネイティブは、Provider GUIの需要が閾値を超えた時だけ追加すれば十分。
- これでUXは単純明快、開発コストは最小、成長余地は確保できます。