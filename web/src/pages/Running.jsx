import React, {useEffect, useState} from "react";
import {Box, Button, ButtonGroup, Chip, FormControl, InputLabel, MenuItem, Select} from "@mui/material";
import {init, pauseMacros, startMacros, stopMacros} from "../api.js";
import useWebSocket, {ReadyState} from "react-use-websocket";

export const Running = (props) => {
    const {value, index, profileName, currentPid, setCurrentPid, ...other} = props;
    const [runningMacrosState, setRunningMacrosState] = useState({});
    const [disabledStart, setDisabledStart] = useState(false);
    const [pidsData, setPidData] = useState([]);
    const [lastMessage, setLastMessage] = useState('');

    useWebSocket(`ws://${import.meta.env.VITE_SERVER_DOMAIN}:${import.meta.env.VITE_SERVER_PORT}/ws`, {
        onOpen: () => console.log('Connected!'),
        onClose: () => console.log('Disconnected!'),
        shouldReconnect: () => true,
        onMessage: (message) => {
            const parsedMessages = JSON.parse(message.data)
            parsedMessages.forEach(item => {
                const parsedMessage = JSON.parse(item)
                for (const messageType in parsedMessage) {
                    if (messageType != 2) {
                        continue;
                    }
                    if (parsedMessage[messageType][currentPid] !== undefined) {
                        let dt = new Date(parsedMessage[messageType][currentPid]['CP']['LastUpdate']);
                        parsedMessage[messageType][currentPid]['CP']['LastUpdate'] = dt.toLocaleTimeString();
                        dt = new Date(parsedMessage[messageType][currentPid]['HP']['LastUpdate']);
                        parsedMessage[messageType][currentPid]['HP']['LastUpdate'] = dt.toLocaleTimeString();
                        dt = new Date(parsedMessage[messageType][currentPid]['MP']['LastUpdate']);

                        parsedMessage[messageType][currentPid]['MP']['LastUpdate'] = dt.toLocaleTimeString();
                        dt = new Date(parsedMessage[messageType][currentPid]['Target']['LastUpdate']);

                        parsedMessage[messageType][currentPid]['Target']['LastUpdate'] = dt.toLocaleTimeString();
                        dt = new Date(parsedMessage[messageType][currentPid]['Target']['HpWasPresentAt']);

                        parsedMessage[messageType][currentPid]['Target']['HpWasPresentAt'] = dt.toLocaleTimeString();
                        dt = new Date(parsedMessage[messageType][currentPid]['Target']['FullHpUnchangedSince']);

                        parsedMessage[messageType][currentPid]['Target']['FullHpUnchangedSince'] = dt.toLocaleTimeString();
                        setLastMessage(JSON.stringify(parsedMessage[messageType][currentPid], null, 2).replace(/ /g, '&nbsp;').replace(/\n/g, '<br>'));
                    }
                }
            })
        }
    });

    const startMacrosAction = async () => {
        setDisabledStart(true);
        try {
            await startMacros(profileName, parseInt(currentPid));
        } catch (e) {
            console.error('startMacros failed', e);
            setDisabledStart(false);
        }
    }
    const pauseMacrosAction = async () => {
        try {
            await pauseMacros(parseInt(currentPid));
        } catch (e) {
            console.error('pauseMacros failed', e);
        }
    }
    const stopMacrosAction = async (pid) => {
        try {
            await stopMacros(pid);
        } catch (e) {
            console.error('stopMacros failed', e);
        } finally {
            setDisabledStart(!currentPid);
        }
    }

    useEffect(() => {
        init().then(({data: {runningMacrosState = {}, PidsData: pidsData}}) => {
            setRunningMacrosState(runningMacrosState);
            setPidData(pidsData);
        }).catch(e => {
            console.log('init failed', e);
            setDisabledStart(true);
        })
    }, []);

    if (value !== index) {
        return null;
    }
    if (!profileName) {
        return null;
    }

    return (
        <Box sx={{m: 2}}>
            <Chip label={profileName} />
            <ButtonGroup variant="contained" sx={{gap: 4, display: 'flex', justifyContent: 'center'}}>
                <FormControl sx={{'width': '200px'}}>
                    <InputLabel id={'pid-label'}>Pid</InputLabel>
                    <Select
                        labelId="pid-label"
                        value={currentPid || ''}
                        label="Pid"
                        onChange={(event) => {
                            setCurrentPid(event.target.value);
                        }}
                        >
                        {Object.keys(pidsData || {}).map((index) => <MenuItem key={index}
                                                                        value={index}>{`${index} - ${pidsData[index]}`}</MenuItem>)}
                    </Select>
                </FormControl>
                <Button color={'error'} onClick={() => stopMacrosAction(parseInt(currentPid))}
                        disabled={runningMacrosState[currentPid]}>Stop</Button>
                <Button onClick={startMacrosAction} disabled={disabledStart}>Start</Button>
                <Button onClick={pauseMacrosAction} disabled={!disabledStart} color={'success'}>Pause</Button>
            </ButtonGroup>
            <br/>
            <div dangerouslySetInnerHTML={{ __html:lastMessage }} />
        </Box>
    );
}
