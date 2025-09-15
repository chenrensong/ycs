import './styles/index.css';
import 'bootstrap/dist/css/bootstrap.min.css';
import React from 'react';
import ReactDOM from 'react-dom';
import { App } from './app';
import { YjsContextProvider } from './context/yjsContext';
import reportWebVitals from './util/reportWebVitals';

const rootElement = document.getElementById('root');

ReactDOM.render(
  <YjsContextProvider wsUrl={'ws://localhost:8080/ws'}>
    <React.StrictMode>
      <App />
    </React.StrictMode>
  </YjsContextProvider>,
  rootElement
);

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals();
