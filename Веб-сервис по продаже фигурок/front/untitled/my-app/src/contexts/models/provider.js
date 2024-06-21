import { useEffect, useReducer } from "react";

import { ModelsContext } from "./context";
import {reducer, initialState} from "./reducer"

export const ModelsProvider = ({ children }) => {
  const [users, dispatch] = useReducer(reducer, initialState)
  return (
    <ModelsContext.Provider value={[users, dispatch]}>
      {children}
    </ModelsContext.Provider>
  );
};

export function GetModels() {
  const [state, dispatch] = useReducer(reducer, initialState)

  useEffect(() => {
    fetch('http://127.0.0.1:8000/models/')
        .then(response => response.json())
        .then(data => {
          dispatch({type: 'GET_DATA', payload: data});
        })
  }, [])
  return state.models
}

export function GetModel(modelId) {
    const [state, dispatch] = useReducer(reducer, initialState)

    useEffect(() => {
        fetch(`http://127.0.0.1:8000/models/${modelId}`)
            .then(response => response.json())
            .then(data => {
                dispatch({type: 'GET_MODEL', payload: data});
            })
    }, [])
    return state.models
}

export function GetCart(user) {
    const [state, dispatch] = useReducer(reducer, initialState)

    useEffect(() => {
        fetch(`http://127.0.0.1:8000/cart/?id_user=${user}`)
            .then(response => response.json())
            .then(data => {
                dispatch({type: 'GET_CART', payload: data});
            })
    }, [])
    return state.cart
}

export function GetPurchases(status) {
    const [state, dispatch] = useReducer(reducer, initialState)

    useEffect(() => {
        fetch(`http://127.0.0.1:8000/sells/?status=${status}`)
            .then(response => response.json())
            .then(data => {
                dispatch({type: 'GET_PURCHASES', payload: data});
            })
    }, [])
    return state.purchases
}

export function GetPurchase(user) {
    const [state, dispatch] = useReducer(reducer, initialState)

    useEffect(() => {
        fetch(`http://127.0.0.1:8000/sells/?id_user=${user}`)
            .then(response => response.json())
            .then(data => {
                dispatch({type: 'GET_PURCHASE', payload: data});
            })
    }, [])
    return state.purchases
}

export function GetBuys() {
    const [state, dispatch] = useReducer(reducer, initialState)

    useEffect(() => {
        fetch(`http://127.0.0.1:8000/status_info/`, {
        })
            .then(response => response.json())
            .then(data => {
                dispatch({type: 'GET_BUYS', payload: data[0]});
            })
    }, [])
    return state.buys
}


