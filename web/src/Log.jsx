import {useEffect, useRef, useState} from "react";
import {Paper, Typography} from "@mui/material";
import useWebSocket, {ReadyState} from "react-use-websocket";
import {throttle} from "lodash/function.js";

export const Log = ({profileName}) => {
    const {readyState} = useWebSocket(`ws://${import.meta.env.VITE_SERVER_DOMAIN}:${import.meta.env.VITE_SERVER_PORT}/ws`, {
        onOpen: () => console.log('Connected!'),
        onClose: () => console.log('Disconnected!'),
        shouldReconnect: () => true,
        disableJson: true,
        onMessage: (message) => setMessages((prev) => [...prev, message.data])
    });
    const messageEndRef = useRef(null);
    const [messages, setMessages] = useState([]);

    useEffect(() => {
        throttle(() => {
            messageEndRef.current?.scrollIntoView({behavior: 'instant'});
        }, 100000)
    }, [messages]);
    const connectionStatus = {
        [ReadyState.CONNECTING]: 'Connecting',
        [ReadyState.OPEN]: 'Open',
        [ReadyState.CLOSING]: 'Closing',
        [ReadyState.CLOSED]: 'Closed',
        [ReadyState.UNINSTANTIATED]: 'Uninstantiated',
    }[readyState];
    if (messages.length >= 200) {
        setMessages(messages.slice(-199));
    }
    return <Paper
        elevation={3}
        sx={{
            flexGrow: 1,
            maxHeight: '100vh',
            overflowY: 'auto',
            p: 2,
            m: 2,
            bgcolor: 'grey.900',
            fontFamily: 'monospace',
        }}
    >
        {messages.map((msg, idx) => (
            <Typography
                className={"log"}
                key={idx} dangerouslySetInnerHTML={{__html: msg}}/>
        ))}
        <div ref={messageEndRef}/>
    </Paper>;
}
