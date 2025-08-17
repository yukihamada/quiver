# QUIVer Node GUI (将来実装)

このディレクトリは、将来的にProvider向けGUIアプリケーションを実装する際の場所です。

## 実装タイミング

以下の条件を2つ以上満たした場合に実装を検討：

1. Providerの非CLI比率 > 40%
2. サポート起因のオンボード工数が月30時間超
3. 家庭用GPU/CPU Providerの日次アクティブが1,000超
4. 収益の>30%が"家庭Provider"から発生

## 現在の推奨

- **Linux/サーバー**: `provider` CLI を使用
- **Windows/Mac**: `provider` CLI をサービス/デーモンとして実行
- **管理・監視**: Web Console (dashboard.quiver.network) を使用

## 参考

- [Provider CLI ドキュメント](../../README.md)
- [Web Console](https://dashboard.quiver.network)