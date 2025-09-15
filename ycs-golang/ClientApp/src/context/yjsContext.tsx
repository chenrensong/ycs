import * as React from 'react';
import * as Y from 'yjs';
import { YjsWebSocketConnector } from '../impl/yjsWebSocketConnector';

export interface IYjsContext {
  readonly yDoc: Y.Doc;
  readonly yjsConnector: YjsWebSocketConnector;
}

export interface IOptions extends React.PropsWithChildren<{}> {
  readonly wsUrl: string;
}

export const YjsContextProvider: React.FunctionComponent<IOptions> = (props: IOptions) => {
  const { wsUrl } = props;

  const contextProps: IYjsContext = React.useMemo(() => {
    const yDoc = new Y.Doc();
    const yjsConnector = new YjsWebSocketConnector(yDoc, wsUrl);

    return { yDoc, yjsConnector };
  }, [wsUrl]);

  // Cleanup on unmount
  React.useEffect(() => {
    return () => {
      contextProps.yjsConnector.destroy();
    };
  }, [contextProps.yjsConnector]);

  return <YjsContext.Provider value={contextProps}>{props.children}</YjsContext.Provider>;
};

export const YjsContext = React.createContext<IYjsContext | undefined>(undefined);
