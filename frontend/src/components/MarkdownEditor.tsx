'use client';

import React from 'react';
import dynamic from 'next/dynamic';
import '@mdxeditor/editor/style.css';

interface MarkdownEditorProps {
  value: string;
  onChange: (value: string) => void;
}

// Dynamically import the entire MDXEditor with all plugins to avoid SSR issues
const MDXEditorComponent = dynamic(
  () => import('@mdxeditor/editor').then((mod) => {
    const {
      MDXEditor,
      headingsPlugin,
      listsPlugin,
      quotePlugin,
      thematicBreakPlugin,
      markdownShortcutPlugin,
      linkPlugin,
      imagePlugin,
      tablePlugin,
      codeBlockPlugin,
      frontmatterPlugin,
      diffSourcePlugin,
      toolbarPlugin,
      UndoRedo,
      BoldItalicUnderlineToggles,
      CodeToggle,
      CreateLink,
      InsertImage,
      InsertTable,
      InsertThematicBreak,
      ListsToggle,
      BlockTypeSelect,
      Separator
    } = mod;

    // Return a component that uses the imported modules
    return function MarkdownEditorInner({ value, onChange }: MarkdownEditorProps) {
      return (
        <div className="markdown-editor border rounded-lg overflow-hidden">
          <MDXEditor
            markdown={value}
            onChange={onChange}
            plugins={[
              // Core plugins
              headingsPlugin(),
              listsPlugin(),
              quotePlugin(),
              thematicBreakPlugin(),
              markdownShortcutPlugin(),
              
              // Enhanced features
              linkPlugin(),
              imagePlugin(),
              tablePlugin(),
              codeBlockPlugin({ defaultCodeBlockLanguage: 'text' }),
              
              // Toolbar
              toolbarPlugin({
                toolbarContents: () => (
                  <>
                    <UndoRedo />
                    <Separator />
                    <BoldItalicUnderlineToggles />
                    <CodeToggle />
                    <Separator />
                    <BlockTypeSelect />
                    <Separator />
                    <CreateLink />
                    <InsertImage />
                    <Separator />
                    <InsertTable />
                    <InsertThematicBreak />
                    <Separator />
                    <ListsToggle />
                  </>
                )
              }),
              
              // Advanced features
              frontmatterPlugin(),
              diffSourcePlugin({ viewMode: 'rich-text', diffMarkdown: '' })
            ]}
            contentEditableClassName="prose dark:prose-invert prose-sm sm:prose lg:prose-lg xl:prose-xl mx-auto focus:outline-none min-h-[200px] p-4"
          />
        </div>
      );
    };
  }),
  {
    ssr: false,
    loading: () => <div className="h-64 flex items-center justify-center text-muted-foreground border rounded-lg">Loading editor...</div>
  }
);

export function MarkdownEditor({ value, onChange }: MarkdownEditorProps) {
  return <MDXEditorComponent value={value} onChange={onChange} />;
}
