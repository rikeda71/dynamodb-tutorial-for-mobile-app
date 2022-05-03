# Memo

## テーブル作成時の注意点

- 異なるエンティティを 1 つのテーブルに保存しているので、プライマリーキーは、UserId のような名前の属性を使用することはできません。この属性は、保存するエンティティのタイプに基づいて異なる何かを意味します。たとえば、ユーザーのプライマリキーは USERNAME であり、ゲームのプライマリキーはその TYPE であるかもしれません。したがって、PK (パーティションキー用) や SK (ソートキー用) などの一般的な名前を属性に使用します。
  - [link](https://aws.amazon.com/jp/getting-started/hands-on/design-a-database-for-a-mobile-app-with-dynamodb/4/) より

## ソートキーの設計

- 文字列タイプのソートキーは、ASCII 文字コードでソートされます。ASCII のドル記号 ($) はポンド記号 (#) の直後にあるため、 Photo エンティティですべてのマッピングを確実に取得できます。
  - [link](https://aws.amazon.com/jp/getting-started/hands-on/design-a-database-for-a-mobile-app-with-dynamodb/4/) より
