import '@fontsource/roboto/300.css';
import '@fontsource/roboto/400.css';
import '@fontsource/roboto/500.css';
import '@fontsource/roboto/700.css';
import './App.css'
import CssBaseline from '@mui/material/CssBaseline';
import {Button, ButtonGroup, createTheme, Grid, ThemeProvider} from "@mui/material";
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
    const startMacrosAction = () => {
        setDisabledStart(true);
        const stFunc = async () => {
            await startMacros(profile);
        }
        try {
            stFunc();
        } finally {
        }
    }
    const stopMacrosAction = () => {
        const stFunc = async () => {
            await stopMacros();
        }
        try {
            stFunc();
        } finally {
            setDisabledStart(false);
        }
    }
    useEffect(() => {
        init().then(({data: {isMacrosRunning, profilesList}}) => {
            console.log('isMacrosRunning', isMacrosRunning);
            setProfilesList(profilesList);
            setDisabledStart(isMacrosRunning);
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
                    <Profiles profilesList={profilesList} setProfile={setProfile} setProfilesList={setProfilesList} />
                </Grid>
                <Grid md={6} xs={12} sx={{mb: 2 }}><Log profileName={profile}/>
                    <ButtonGroup variant="contained" sx={{gap: 4, display: 'flex', justifyContent: 'center'}}>
                        <Button color={'error'} onClick={stopMacrosAction} disabled={!disabledStart}>Stop</Button>
                        <Button onClick={startMacrosAction} disabled={disabledStart}>Start</Button>
                    </ButtonGroup>
                </Grid>
            </Grid>
        </ThemeProvider>
    )
}

export default App
