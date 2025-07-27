'use client';

import React, { useState } from 'react';
import ReactMde from 'react-mde';
import Showdown from 'showdown';
import 'react-mde/lib/styles/css/react-mde-all.css';

interface MarkdownEditorProps {
  value: string;
  onChange: (value: string) => void;
}

export function MarkdownEditor({ value, onChange }: MarkdownEditorProps) {
  const [selectedTab, setSelectedTab] = useState<'write' | 'preview'>('write');

  const converter = new Showdown.Converter({
    tables: true,
    simplifiedAutoLink: true,
  });

  return (
    <div className="markdown-editor">
      <ReactMde
        value={value}
        onChange={onChange}
        selectedTab={selectedTab}
        onTabChange={setSelectedTab}
        generateMarkdownPreview={(markdown) =>
          Promise.resolve(converter.makeHtml(markdown))
        }
      />
    </div>
  );
}
