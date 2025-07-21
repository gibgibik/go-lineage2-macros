import '@fontsource/roboto/300.css';
import '@fontsource/roboto/400.css';
import '@fontsource/roboto/500.css';
import '@fontsource/roboto/700.css';
import './App.css'
import CssBaseline from '@mui/material/CssBaseline';
import {
    Button,
    ButtonGroup,
    createTheme,
    FormControl,
    Grid,
    InputLabel,
    MenuItem,
    Select,
    ThemeProvider
} from "@mui/material";
import {Log} from "./Log.jsx";
import {Macros} from "./Macros.jsx";
import {useEffect, useState} from "react";
import {init, startMacros, stopMacros} from "./api.js";
import {Profiles} from "./Profile.jsx";

const theme = createTheme({
    palette: {
        mode: 'dark',  // 'light' or 'dark'
        primary: {
            main: '#1976d2',
        },
    },
});

// const PROFILE_NAME = 'static_profile'

function App() {
    const [disabledStart, setDisabledStart] = useState(false);
    const [profile, setProfile] = useState(null);
    const [profilesList, setProfilesList] = useState([]);
    const [pidsData, setPidData] = useState([]);
    const [currentPid, setCurrentPid] = useState(null);
    const [runningMacrosState, setRunningMacrosState] = useState({});
    const startMacrosAction = () => {
        setDisabledStart(true);
        const stFunc = async () => {
            await startMacros(profile, parseInt(currentPid));
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
            setDisabledStart(!currentPid || !profile);
        }
    }
    useEffect(() => {
        setDisabledStart(!profile || !runningMacrosState[currentPid]);
    }, [profile, currentPid]);
    useEffect(() => {
        init().then(({data: {runningMacrosState = {}, profilesList, PidsData: pidsData}}) => {
            setProfilesList(profilesList);
            setRunningMacrosState(runningMacrosState);
            setDisabledStart(!profile || !runningMacrosState[currentPid]);
            setPidData(pidsData);
        }).catch(e => {
            console.log('init failed', e);
            setDisabledStart(true);
        })
    }, []);
    return (
        <ThemeProvider theme={theme}>
            <CssBaseline/>
            <Grid container spacing={0} sx={{flexDirection: 'column', width: '100vw'}}>
                <Grid md={6} xs={12} sx={{maxWidth: '100vw', display: 'flex'}}>
                    <Macros profileName={profile}/>
                    <Profiles profilesList={profilesList} setProfile={setProfile} setProfilesList={setProfilesList}/>
                </Grid>
                <Grid md={6} xs={12} sx={{mb: 2}}><Log profileName={profile}/>
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
                        <Button color={'error'} onClick={() => stopMacrosAction(parseInt(currentPid))} disabled={runningMacrosState[currentPid]}>Stop</Button>
                        <Button onClick={startMacrosAction} disabled={disabledStart} >Start</Button>
                    </ButtonGroup>
                </Grid>
            </Grid>
        </ThemeProvider>
    )
}

export default App
