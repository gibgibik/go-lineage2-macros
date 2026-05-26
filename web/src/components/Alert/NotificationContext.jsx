import {createContext, useState} from "react";
import Notification from "./Alert.jsx";


export const NotificationContext = createContext();

export const AlertProvider = ({children}) => {
    const [notification, setNotification] = useState(null);
    const setAlert = (text) => {
        setNotification({ text, severity: 'error', id: Date.now() });
    };

    const setSuccess = (text) => {
        setNotification({ text, severity: 'success', id: Date.now() });
    };
    return (
        <NotificationContext.Provider value={{setAlert, setSuccess}}>
            {notification && <Notification key={notification.id}  text={notification.text}
                                           severity={notification.severity}/>}
            {children}
        </NotificationContext.Provider>
    );
}
