import React from 'react';
import * as Y from 'yjs';
import { YjsContext } from '../context/yjsContext';
import { YjsWebSocketConnector } from '../impl/yjsWebSocketConnector';

export function useYjs(): { yDoc: Y.Doc, yjsConnector: YjsWebSocketConnector } {
  const yjsContext = React.useContext(YjsContext);
  if (yjsContext === undefined) {
    throw new Error('useYjs() should be called with the YjsContext defined.');
  }

  return {
    yDoc: yjsContext.yDoc,
    yjsConnector: yjsContext.yjsConnector
  };
}
