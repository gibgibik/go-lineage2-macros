import React, {useEffect, useState} from "react";
import {Box, Button, ButtonGroup, Chip, FormControl, InputLabel, MenuItem, Select} from "@mui/material";
import {init, pauseMacros, startMacros, stopMacros} from "../api.js";
import useWebSocket, {ReadyState} from "react-use-websocket";

export const Running = (props) => {
    const {value, index, profileName, currentPid, setCurrentPid} = props;
    const [runningMacrosState, setRunningMacrosState] = useState({});
    const [disabledStart, setDisabledStart] = useState(false);
    const [pidsData, setPidData] = useState([]);
    useWebSocket(`ws://${import.meta.env.VITE_SERVER_DOMAIN}:${import.meta.env.VITE_SERVER_PORT}/ws`, {
        onOpen: () => console.log('Connected!'),
        onClose: () => console.log('Disconnected!'),
        shouldReconnect: () => true,
        // disableJson: false,
        onMessage: (message) => {
            const parsedMessages = JSON.parse(message.data)
            parsedMessages.forEach(item => {
                console.log(item);
            })
        }
    });
    useEffect(() => {
        init().then(({data: {runningMacrosState = {}, PidsData: pidsData}}) => {
            setRunningMacrosState(runningMacrosState);
            // setDisabledStart(!runningMacrosState[currentPid]);
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
        return;
    }
    const startMacrosAction = () => {
        setDisabledStart(true);
        const stFunc = async () => {
            await startMacros(profileName, parseInt(currentPid));
        }
        stFunc();
    }
    const pauseMacrosAction = () => {
        const stFunc = async () => {
            await pauseMacros(parseInt(currentPid));
        }
        stFunc();
    }
    const stopMacrosAction = (pid) => {
        const stFunc = async () => {
            await stopMacros(pid);
        }
        try {
            stFunc();
        } finally {
            setDisabledStart(!currentPid);
        }
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
        </Box>
    );
}
