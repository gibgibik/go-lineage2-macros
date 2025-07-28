import React, {useEffect, useState} from "react";
import {Box, Button, ButtonGroup, Chip, FormControl, InputLabel, MenuItem, Select} from "@mui/material";
import {init, pauseMacros, startMacros, stopMacros} from "../api.js";

export const Running = (props) => {
    const {value, index, profileName, currentPid, setCurrentPid, ...other} = props;
    if (value !== index) {
        return null;
    }
    if (!profileName) {
        return;
    }
    const [runningMacrosState, setRunningMacrosState] = useState({});
    const [disabledStart, setDisabledStart] = useState(false);
    const [pidsData, setPidData] = useState([]);
    const startMacrosAction = () => {
        setDisabledStart(true);
        const stFunc = async () => {
            await startMacros(profileName, parseInt(currentPid));
        }
        try {
            stFunc();
        } finally {
        }
    }
    const pauseMacrosAction = () => {
        const stFunc = async () => {
            await pauseMacros(parseInt(currentPid));
        }
        try {
            stFunc();
        } finally {
        }
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
    useEffect(() => {
        init().then(({data: {runningMacrosState = {}, profilesList, PidsData: pidsData}}) => {
            setRunningMacrosState(runningMacrosState);
            console.log(runningMacrosState);
            // setDisabledStart(!runningMacrosState[currentPid]);
            setPidData(pidsData);
        }).catch(e => {
            console.log('init failed', e);
            setDisabledStart(true);
        })
    }, []);
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
                        {Object.keys(pidsData).map((index) => <MenuItem key={index}
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
