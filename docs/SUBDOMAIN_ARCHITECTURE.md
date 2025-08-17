# QUIVer サブドメイン設計仕様書

## 1. 全体像（役割と置くもの）

```
quiver.network              # マーケ用の玄関（ビジョン / USP / ライブ統計 / CTA）
├─ explorer.quiver.network  # ノード/ジョブの可視化（地図・ランキング・履歴）
├─ docs.quiver.network      # ドキュメント（Provider / Developer / Contributor）
├─ api.quiver.network       # APIポータル（OAS, 認証, 例, SDKリンク）
├─ playground.quiver.network # ブラウザP2Pデモ（WebRTCで即体験）
├─ dashboard.quiver.network # Provider用ダッシュボード（ノード・報酬・設定）
├─ status.quiver.network    # 稼働状況/インシデント履歴（SLA, メトリクス）
├─ security.quiver.network  # セキュリティ&VDP（脅威モデル/鍵管理/報告窓口）
├─ blog.quiver.network      # アップデート/研究ノート/事例
├─ community.quiver.network # Discord・フォーラム・イベント集約
├─ registry.quiver.network  # モデル/バイナリ/Model Packの署名配布（Sigstore等）
└─ cdn.quiver.network       # 静的配布（モデルカード画像/JS/CSS/ドキュメント資材）
```

## 2. QuicPair との関係
- QuicPairのブランドは独立のまま（例：quicpair.app）
- quiver.network/quicpair は連携案内のランディング（"Private first → Scale when you choose"）
- QuicPairアプリからQUIVerを使う際の技術文書は docs.quiver.network/integrations/quicpair に集約

## 3. 各サブドメインの詳細仕様

### quiver.network（ルート）
**目的**: 初見で理解→信頼→行動（Provider登録 / Dev導線 / デモ）

**コンテンツ構成**:
1. **ヒーロー（1画面で完結）**
   - キャッチ：「世界の遊休リソースを束ねる、検証可能な分散AIネットワーク」
   - サブ：Private first. Scale when you choose.
   - ライブ統計（WebSocket/SSEで更新）：アクティブノード数 / 推論/秒 / 総TFLOPS / 参加国
   - CTA：「Providerになる」 / 「APIで使う」 / 「ライブデモ」

2. **なぜ必要か（3カード）**
   - 民主化（集中回避）
   - 検証可能性（署名レシート）
   - コストとレイテンシ（近接処理）

3. **How it works（図1枚）**
   - クライアント →（P2P/中継）→ Provider
   - 計算レシート（Ed25519署名+Merkle）→ 検証
   - プライバシー前提/ ポリシー付きジョブ

4. **ユースケース**
   - 画像生成・長文要約・夜間バッチ・マルチモーダル等

5. **信頼セクション**
   - 100% OSS / 署名済みリリース / 再現ビルド / セキュリティポリシー / VDP

6. **導線（3ボタン）**
   - Provider ▶ dashboard.
   - Developer ▶ docs.
   - デモ ▶ playground.

7. **フッター**
   - 法務（利用規約/プライバシー）
   - セキュリティ.txt
   - ステータス
   - ブログ
   - コミュニティ

### explorer.quiver.network
**目的**: 透明性の可視化とワクワク感

**機能**:
- 地図＋ヒートマップ（国/リージョン/ISPは粗い粒度で）
- ノード一覧（匿名ID、スペック、稼働率、対応モデル）
- ジョブフィード（モデルID、実行時間、署名ハッシュ、料金、結果サイズ）
- ランキング（稼働時間/完了ジョブ/評価）
- 検索/フィルタ（地域・モデル・価格・GPU種別）
- 免責（位置は概算、プライバシー保護により詳細は非公開）

### docs.quiver.network
**3本柱で即スタート**:

1. **Provider（稼ぐ人）**
   - 要件（CPU/GPU/RAM/電源）
   - インストール（macOS PKG / Linux Docker / CLI）
   - 初回起動→ノード登録→テストジョブ
   - ポリシー（稼働時間/温度/上限電力/自動停止）
   - 報酬（計算式、支払いフロー、手数料、税務の一般的注意）

2. **Developer（使う人）**
   - 5分クイックスタート（JS/Goサンプル）
   - API仕様（OpenAPI）/ 認証 / レート
   - 計算レシート検証（署名/ハッシュの検証コード）
   - 料金の見積 & 上限設定（事前見積→同意→実行）

3. **Contributor（貢献する人）**
   - 開発手順 / テスト / セキュリティ方針 / コード規約
   - ロードマップ / Good first issue

**統合**: docs./integrations/quicpair に QuicPair 連携ガイド

### api.quiver.network
- OpenAPI（Swagger UI）のホスト
- サンプルキー発行（テスト用）
- SDKリンク（JS/TS, Go, Python）
- レートと課金のリファレンス（従量メトリクス／見積API／上限設定）

### playground.quiver.network
- ブラウザで即P2P推論（WebRTC）
- モデル選択（軽量テキスト→まず成功体験）
- レイテンシと概算コストを実行前に表示
- 実行後にレシートを展開表示

### dashboard.quiver.network
- Providerログイン（Wallet/Email/SIWAなど）
- ノード管理（稼働/温度/VRAM/待機）
- 収益と出金（履歴・手数料・明細）
- ポリシー（時間帯・電力・地域制限）
- 通知（温度超過、クラッシュ、利用規約更新）

### status.quiver.network
- 稼働状況、過去インシデント、稼働率
- コンポーネント別（API/Registry/Explorer）
- Webhook/Atom/RSS

### security.quiver.network
- 脅威モデル・暗号スイート（Noise IK / X25519 / AES‑GCM / BLAKE2）
- データ取扱い（転送暗号化・ノード上の扱い・オプションのTEE/分割実行）
- VDP（報告窓口、報奨の有無、SLA）
- 署名キー/PGP、再現ビルド手順
- /.well-known/security.txt（ルートにも配置）

### blog.quiver.network
- リリース、研究メモ、エコシステム採用事例

### community.quiver.network
- Discord, フォーラム, ガバナンス告知

### registry.quiver.network
- 署名済みModel Pack / クライアントの配布（ハッシュ/署名/SBOM）

### cdn.quiver.network
- モデルカード画像、静的JS/CSS、バイナリのキャッシュ

## 4. SEO / i18n / 計測（プライバシー配慮）

- **i18n**: /ja と /en の2言語。既定はブラウザロケール
- **メタ/構造化**: OG/Twitter Card、JSON‑LD（SoftwareApplication, Dataset など）
- **サイトマップ**: /sitemap.xml、/robots.txt
- **計測**: Plausible/Umami等のクッキー不要アナリティクス（IP匿名化）
- **フォーム**: 問い合わせはEmailリンク or GitHub Issuesへ（個人情報最小化）

## 5. 運用・セキュア設定（チェックリスト）

- **DNS**: DNSSEC / CAA（Let's Encrypt/Vendor）
- **TLS**: 1.3/HSTS/OCSP stapling
- **HTTPヘッダ**: CSP, X-Content-Type-Options, Referrer-Policy, Permissions-Policy
- **バイナリ/パック**: Sigstore（またはGPG）署名、SBOM、ハッシュ掲載
- **バックアップ**: Registry/Blog/DocsはGitで版管理
- **アクセシビリティ**: WCAG 2.1 AA 目標（コントラスト、キーボード操作）

## 6. 最小公開セット（まず置くと良いもの）

1. **quiver.network**：ヒーロー＋ライブ統計＋3つの導線（Provider/Dev/デモ）
2. **docs.**：Provider/Developerのクイックスタート各1ページ
3. **playground.**：軽量モデルでのテキスト推論デモ
4. **security.**：脅威モデル概要＋VDP＋security.txt
5. **explorer.**：地図と簡易ランキング（ダミーデータ→本番移行）

## 7. 技術スタック（提案）

- **フロント**: Next.js（i18n/ISR）
- **Docs**: Docusaurus（バージョニング）
- **Explorer**: Next.js + WebSocket/SSE + 地図（MapLibre）
- **Playground**: WebRTC + Wasm/Worker
- **配信**: Cloudflare Pages/Workers or Vercel、静的は R2/CDN
- **署名配布**: Sigstore + GitHub Releases（registry. で配布）

## 8. ディレクトリ構造

```
/Users/yuki/QUIVer/docs/
├── index.html                  # quiver.network（ルート）
├── quicpair/                   # QuicPair連携ページ
├── explorer/                   # explorer.quiver.network
│   ├── index.html
│   └── js/
├── api/                        # api.quiver.network
│   ├── index.html
│   └── openapi.yaml
├── playground/                 # playground.quiver.network
│   ├── index.html
│   └── playground-stream.html
├── dashboard/                  # dashboard.quiver.network
├── status/                     # status.quiver.network
├── security/                   # security.quiver.network
│   ├── index.html
│   └── .well-known/
│       └── security.txt
├── blog/                       # blog.quiver.network
├── community/                  # community.quiver.network
├── registry/                   # registry.quiver.network
└── cdn/                        # cdn.quiver.network
```

## 9. 実装順序

1. **Phase 1（即日）**
   - quiver.network の新トップページ
   - playground.quiver.network（既存のplayground-stream.html移行）
   - security.quiver.network（基本情報）

2. **Phase 2（1週間）**
   - explorer.quiver.network（地図とダミーデータ）
   - docs.quiver.network（Docusaurus初期設定）
   - api.quiver.network（OpenAPIドキュメント）

3. **Phase 3（2週間）**
   - dashboard.quiver.network（Provider管理画面）
   - status.quiver.network（稼働状況）
   - QuicPair統合ページ

4. **Phase 4（随時）**
   - blog.quiver.network
   - community.quiver.network
   - registry.quiver.network
   - cdn.quiver.network