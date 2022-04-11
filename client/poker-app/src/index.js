import React from 'react'
import ReactDOM from 'react-dom'

import './index.css'

import App from './App'
import { AppStateProvider } from './appStore'
import reportWebVitals from './reportWebVitals'
import { WebSocketProvider } from './WebSocket'
//################# 22-04-11 #################
// yarn add @web3-react/core @ethersproject/providers
// yarn add @web3-react/injected-connector
import { Web3ReactProvider } from "@web3-react/core"
import { Web3Provider } from "@ethersproject/providers"
function getLibrary(provider) {
  const library = new Web3Provider(provider, "any")
  return library
}

ReactDOM.render(
  <React.StrictMode>
    <AppStateProvider>
      <WebSocketProvider>
        <Web3ReactProvider getLibrary={getLibrary}>
          <App />
        </Web3ReactProvider>
      </WebSocketProvider>
    </AppStateProvider>
  </React.StrictMode>,
  document.getElementById('root')
)

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals(console.log)
