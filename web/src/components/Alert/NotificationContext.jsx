import {createContext, useState} from "react";
import Notification from "./Alert.jsx";


export const NotificationContext = createContext();

export const AlertProvider = ({children}) => {
    const [text, setText] = useState(null);
    const [severity, setSeverity] = useState(null);
    const setAlert = (text) => {
        setText(text)
        setSeverity('error');
    }
    const setSuccess = (text) => {
        setText(text)
        setSeverity('success');
    }
    return (
        <NotificationContext.Provider value={{setAlert, setSuccess}}>
            {text && severity && <Notification text={text} severity={severity}/>}
            {children}
        </NotificationContext.Provider>
    );
}
