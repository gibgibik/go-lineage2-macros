import '@fontsource/roboto/300.css';
import '@fontsource/roboto/400.css';
import '@fontsource/roboto/500.css';
import '@fontsource/roboto/700.css';
import './App.css'
import CssBaseline from '@mui/material/CssBaseline';
import {Button, ButtonGroup, createTheme, Grid, ThemeProvider} from "@mui/material";
import {Log} from "./Log.jsx";
import {Macros} from "./Macros.jsx";
import {useState} from "react";
import {startMacros, stopMacros} from "./api.js";

const theme = createTheme({
    palette: {
        mode: 'dark',  // 'light' or 'dark'
        primary: {
            main: '#1976d2',
        },
    },
});

const PROFILE_NAME = 'static_profile'

function App() {
    const [disabledStart, setDisabledStart] = useState(false);
    const [disabledStop, setDisabledStop] = useState(true);
    const startMacrosAction = () => {
        setDisabledStart(true);
        const stFunc = async () => {
            await startMacros(PROFILE_NAME);
        }
        try {
            stFunc();
        } finally {
            setDisabledStop(false);
        }
    }
    const stopMacrosAction = () => {
        const stFunc = async () => {
            await stopMacros();
        }
        try {
            stFunc();
        } finally {
            setDisabledStop(true);
            setDisabledStart(false);
        }
    }
    return (
        <ThemeProvider theme={theme}>
            <CssBaseline/>
            <Grid  container spacing={0} sx={{width: '100vw'}}>
                <Grid md={6} xs={12} ><Macros profileName={PROFILE_NAME}/></Grid>
                <Grid md={6} xs={12} sx={{mb: 22}}><Log profileName={PROFILE_NAME}/>
                    <ButtonGroup variant="contained" sx={{gap: 4, display: 'flex', justifyContent: 'center'}}>
                        <Button color={'error'} onClick={stopMacrosAction} disabled={disabledStop}>Stop</Button>
                        <Button onClick={startMacrosAction} disabled={disabledStart}>Start</Button>
                    </ButtonGroup>
                </Grid>
            </Grid>
        </ThemeProvider>
    )
}

export default App
