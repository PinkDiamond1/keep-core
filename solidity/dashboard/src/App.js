import React, { useEffect } from "react"
import { Link } from "react-router-dom"
import Web3ContextProvider from "./components/Web3ContextProvider"
import Routing from "./components/Routing"
import { Messages } from "./components/Message"
import { SideMenu } from "./components/SideMenu"
import { BrowserRouter as Router } from "react-router-dom"
import { Provider, useDispatch } from "react-redux"
import store from "./store"
import { ModalContextProvider } from "./components/Modal"
import * as Icons from "./components/Icons"
import Footer from "./components/Footer"
import { usePrevious } from "./hooks/usePrevious"
import { isSameEthAddress } from "./utils/general.utils"
import { useWeb3Context } from "./components/WithWeb3Context"

const App = () => (
  <Provider store={store}>
    <Router basename={`${process.env.PUBLIC_URL}`}>
      <Messages>
        <Web3ContextProvider>
          <ModalContextProvider>
            <AppLayout />
          </ModalContextProvider>
        </Web3ContextProvider>
      </Messages>
    </Router>
  </Provider>
)

const AppLayout = () => {
  const { isConnected, connector, yourAddress } = useWeb3Context()
  const dispatch = useDispatch()

  useEffect(() => {
    const eventHandler = (address) => {
      dispatch({ type: "app/account_changed", payload: { address } })
    }
    if (isConnected) {
      dispatch({ type: "app/set_account", payload: { address: yourAddress } })
      connector.on("accountsChanged", eventHandler)
    }

    return () => {
      if (connector) {
        connector.removeListener("accountsChanged", eventHandler)
      }
    }
  }, [isConnected, connector, dispatch, yourAddress])

  return (
    <>
      <AppHeader />
      <section className="app__content">
        <Routing />
      </section>
      <Footer className="app__footer" />
    </>
  )
}

const AppHeader = () => {
  return (
    <header className="app__header">
      <Link to="/" className="app-logo">
        <Icons.KeepDashboardLogo />
      </Link>
      <SideMenu />
    </header>
  )
}
export default App
