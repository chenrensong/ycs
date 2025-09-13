import * as React from 'react';
import * as yMonaco from 'y-monaco';
import MonacoEditor, { monaco } from 'react-monaco-editor';
import { useYjs } from '../hooks/useYjs';

export const YjsMonacoEditor = () => {
  const { yDoc, yjsConnector } = useYjs();
  const yText = yDoc.getText('monaco');

  const setMonacoEditor = React.useState<monaco.editor.ICodeEditor>()[1];
  const setMonacoBinding = React.useState<yMonaco.MonacoBinding>()[1];

  const _onEditorDidMount = React.useCallback(
    (editor: monaco.editor.ICodeEditor, monacoParam: typeof monaco): void => {
      editor.focus();
      editor.setValue('');

      const model = editor.getModel();
      if (model) {
        setMonacoEditor(editor);
        setMonacoBinding(new yMonaco.MonacoBinding(yText, model as any, new Set([editor as any]), yjsConnector.awareness));
      }
    },
    [yjsConnector.awareness, yText, setMonacoEditor, setMonacoBinding]
  );

  return (
    <div>
      <MonacoEditor
        width='100%'
        height='600px'
        theme='vs'
        options={{
          automaticLayout: true
        }}
        editorDidMount={(e, a) => _onEditorDidMount(e, a)}
      />
    </div>
  );
};
