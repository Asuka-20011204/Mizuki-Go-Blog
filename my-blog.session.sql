INSERT INTO posts (
    id,
    created_at,
    updated_at,
    deleted_at,
    title,
    slug,
    description,
    content,
    category,
    tags,
    image,
    lang,
    pinned
  )
VALUES (
    'id:bigint',
    'created_at:datetime',
    'updated_at:datetime',
    'deleted_at:datetime',
    'title:longtext',
    'slug:varchar',
    'description:longtext',
    'content:longtext',
    'category:longtext',
    'tags:longtext',
    'image:longtext',
    'lang:longtext',
    'pinned:tinyint'
  );ALTER TABLE posts ADD COLUMN description LONGTEXT;