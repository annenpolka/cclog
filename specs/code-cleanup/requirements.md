# Requirements Document - Code Cleanup

## Introduction
このプロジェクトは cclog プロジェクトの不要なコードを削除・整理し、コードベースの保守性と可読性を向上させることを目的としています。現在のコードベースには legacy関数や長い関数名が含まれており、これらを適切に整理・リネームする必要があります。

## Requirements

### Requirement 1: Legacy関数の削除とリネーミング
**User Story:** 開発者として、legacy関数（formatMessage）を削除し、冗長な関数名（formatMessageWithOptions）をより簡潔で明確な名前（formatMessage）にリネームしたい。これにより、コードの一貫性と可読性が向上する。

#### Acceptance Criteria
1. WHEN 開発者がmarkdown.goファイルを確認するとき THEN システムはlegacy formatMessage関数が削除されていることを確認する
2. WHEN 開発者がコードを確認するとき THEN システムはformatMessageWithOptionsがformatMessageにリネームされていることを確認する
3. WHEN 開発者がテストを実行するとき THEN システムは全てのテストが正常に通ることを確認する
4. IF 既存のコードがformatMessageWithOptionsを呼び出している場合 THEN システムは新しいformatMessage関数を使用するように更新する
5. WHEN 開発者がAPIを使用するとき THEN システムはより直感的な関数名でアクセスできることを確認する

#### Additional Considerations
- **Performance**: 既存の性能を維持する
- **Security**: 変更によるセキュリティリスクがないことを確認
- **Usability**: APIがより使いやすくなることを確認

### Requirement 2: 関数名の一貫性向上
**User Story:** 開発者として、プロジェクト全体の関数命名規則を見直し、より一貫性のある命名にしたい。特に"WithOptions"のような冗長なサフィックスを削除し、デフォルト引数パターンを採用したい。

#### Acceptance Criteria
1. WHEN 開発者がコードベースを確認するとき THEN システムは一貫した命名規則が適用されていることを確認する
2. WHEN 開発者が新しい関数を追加するとき THEN システムは既存の命名パターンに従うことができる
3. IF 関数がオプション引数を受け取る場合 THEN システムはデフォルト値を持つシンプルな引数として設計する
4. WHILE コードレビューを行うとき THEN システムは命名の一貫性を保持していることを確認する

### Requirement 3: 未使用コメントの整理
**User Story:** 開発者として、コード内の古いコメントやメモを整理して、現在の実装状況に合わせた適切なドキュメンテーションにしたい。

#### Acceptance Criteria
1. WHEN 開発者がコードを確認するとき THEN システムは古い"Note:"コメントが適切に更新または削除されていることを確認する
2. WHEN 開発者がlegacy関数のコメントを確認するとき THEN システムは"legacy function"という記述が削除されていることを確認する
3. WHEN 開発者がテストファイルを確認するとき THEN システムは説明的でないコメントが削除されていることを確認する
4. IF コメントが実装の説明に重要である場合 THEN システムはコメントを保持し、必要に応じて更新する

### Requirement 4: 不要なテストコードの削除
**User Story:** 開発者として、使用されていないテストヘルパーやテスト用の一時ファイルチェックロジックを削除して、テストコードをより簡潔にしたい。

#### Acceptance Criteria
1. WHEN 開発者がテストファイルを確認するとき THEN システムは重複したos.IsNotExist()チェックが統合されていることを確認する
2. WHEN 開発者がテストを実行するとき THEN システムは削除後も全てのテストが正常に動作することを確認する
3. WHILE テストが実行されているとき THEN システムは適切なテンポラリファイルクリーンアップが実行されることを確認する

#### Additional Considerations
- **Reliability**: テストの信頼性を維持する
- **Performance**: テスト実行時間の改善
- **Maintainability**: テストコードの保守性向上

### Requirement 5: 冗長なwrapper関数の削除
**User Story:** 開発者として、単純にデフォルト引数で他の関数を呼ぶだけのwrapper関数を削除し、引数を使った統一的なAPIにしたい。これにより、関数数が減り、APIがシンプルになる。

#### Acceptance Criteria
1. WHEN 開発者がformatterパッケージを確認するとき THEN システムはFormatConversationToMarkdownがオプション引数を受け取る統一APIになっていることを確認する
2. WHEN 開発者がtypesパッケージを確認するとき THEN システムはTruncateTitleがwidth引数を受け取る統一APIになっていることを確認する
3. WHEN 開発者がextractMessageContent関数を確認するとき THEN システムは単一の関数でオプション制御されていることを確認する
4. IF 既存のコードが削除されたwrapper関数を使用している場合 THEN システムは適切な引数付きで統一関数を呼び出すように更新する

#### Additional Considerations
- **API Consistency**: 全ての関数で一貫したオプション指定パターンを採用
- **Performance**: wrapper関数による不要な関数呼び出しオーバーヘッドを削除
- **Maintainability**: 関数数の削減による保守性向上

### Requirement 6: WithOptionsパターンの統一
**User Story:** 開発者として、同じ機能を持つ2つの関数（デフォルト版とWithOptions版）を1つの関数に統一し、オプション引数でデフォルト動作を制御したい。

#### Acceptance Criteria
1. WHEN 開発者がmarkdownパッケージを確認するとき THEN システムは各機能に対して1つの関数のみが存在することを確認する
2. WHEN 開発者がAPIを使用するとき THEN システムはオプション構造体でデフォルト動作を指定できることを確認する
3. WHILE 後方互換性を保つとき THEN システムは既存の呼び出しパターンが動作することを確認する
4. IF WithOptionsパターンが削除される場合 THEN システムは全ての使用箇所が新しいAPIに更新されることを確認する

### Requirement 7: インポートの最適化
**User Story:** 開発者として、使用されていないインポート文を削除し、必要なインポートのみを保持することで、コンパイル時間を短縮し、依存関係を明確にしたい。

#### Acceptance Criteria
1. WHEN 開発者がgo mod tidyを実行するとき THEN システムは不要な依存関係が削除されることを確認する
2. WHEN 開発者がgoimportsを実行するとき THEN システムは未使用のインポートが削除されることを確認する
3. WHEN 開発者がコードをコンパイルするとき THEN システムは未使用インポートの警告が表示されないことを確認する

## Non-Functional Requirements

### Performance
- コード削除後のビルド時間が現在の時間以下であること
- テスト実行時間が現在の時間以下であること
- 実行時性能に影響がないこと
- 関数名変更による呼び出しオーバーヘッドが発生しないこと

### Security
- セキュリティ関連のコードは削除対象から除外すること
- 認証・認可に関わるコードは慎重に確認すること
- 関数名変更による意図しない動作変更がないこと

### Reliability
- 全てのテストが正常に通ること
- 既存の機能に影響がないこと
- エラーハンドリングが適切に保持されること
- 関数リネーミング後も同じ動作を保証すること

### Maintainability
- コードの可読性が向上すること
- 将来の保守作業が容易になること
- ドキュメンテーションが適切に更新されること
- 新しい開発者が理解しやすい命名規則であること

### Usability
- APIがより直感的になること
- 関数名から機能が推測しやすくなること
- IDEの補完機能でより見つけやすくなること

## Assumptions and Constraints

### Assumptions
- 現在のAPIの機能は変更しない（パラメータと戻り値は同じ）
- 既存のテストカバレッジレベルを維持する
- Go標準ライブラリのみを使用する部分は変更しない
- 関数のシグネチャは互換性を保つ

### Constraints
- 既存の機能を破壊してはならない
- パブリックAPIの動作は変更してはならない
- TDDアプローチに従い、テストファーストで進める
- 既存のプロジェクト構造を保持する
- リネーミングは段階的に行い、必要に応じて一時的にエイリアスを作成する

## Glossary
- **Legacy関数**: 古いバージョンとの互換性のために残されているが、新しい実装で置き換え可能な関数
- **未使用コード**: 現在参照されていない、または実際に使用されていないコード
- **Code Cleanup**: コードベースから不要な要素を削除し、構造を改善するプロセス
- **Refactoring**: 外部的な動作を変更せずに内部構造を改善すること
- **関数リネーミング**: より適切で一貫性のある名前に変更すること
- **WithOptions パターン**: オプション引数を別関数として分離するGoの一般的なパターン