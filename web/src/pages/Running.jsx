import React, {useEffect, useState} from "react";
import {Box, Button, ButtonGroup, Chip, FormControl, InputLabel, MenuItem, Select} from "@mui/material";
import {init, pauseMacros, startMacros, stopMacros} from "../api.js";
import useWebSocket, {ReadyState} from "react-use-websocket";

export const Running = (props) => {
    const {value, index, profileName, currentPid, setCurrentPid, ...other} = props;
    const [runningMacrosState, setRunningMacrosState] = useState({});
    const [pausedState, setPausedState] = useState({});
    const [pidsData, setPidData] = useState([]);
    const [lastMessage, setLastMessage] = useState('');

    // Single source of truth: is the currently-selected pid running / paused?
    const isRunning = !!currentPid && !!runningMacrosState[currentPid];
    const isPaused = !!currentPid && !!pausedState[currentPid];

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
        if (!currentPid) {
            return;
        }
        // Optimistically mark as running for the selected pid.
        setRunningMacrosState(prev => ({...prev, [currentPid]: true}));
        setPausedState(prev => ({...prev, [currentPid]: false}));
        try {
            await startMacros(profileName, parseInt(currentPid));
        } catch (e) {
            console.error('startMacros failed', e);
            setRunningMacrosState(prev => ({...prev, [currentPid]: false}));
        }
    }
    const pauseMacrosAction = async () => {
        if (!currentPid) {
            return;
        }
        // pause endpoint is a toggle on the backend (pause <-> resume).
        try {
            await pauseMacros(parseInt(currentPid));
            setPausedState(prev => ({...prev, [currentPid]: !prev[currentPid]}));
        } catch (e) {
            console.error('pauseMacros failed', e);
        }
    }
    const stopMacrosAction = async (pid) => {
        try {
            await stopMacros(pid);
            setRunningMacrosState(prev => ({...prev, [pid]: false}));
            setPausedState(prev => ({...prev, [pid]: false}));
        } catch (e) {
            console.error('stopMacros failed', e);
        }
    }

    useEffect(() => {
        init().then(({data: {runningMacrosState = {}, PidsData: pidsData}}) => {
            setRunningMacrosState(runningMacrosState);
            setPidData(pidsData);
        }).catch(e => {
            console.log('init failed', e);
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
                        disabled={!isRunning}>Stop</Button>
                <Button onClick={startMacrosAction} disabled={!currentPid || isRunning}>Start</Button>
                <Button onClick={pauseMacrosAction} disabled={!isRunning} color={'success'}>
                    {isPaused ? 'Resume' : 'Pause'}
                </Button>
            </ButtonGroup>
            <br/>
            <div dangerouslySetInnerHTML={{ __html:lastMessage }} />
        </Box>
    );
}
