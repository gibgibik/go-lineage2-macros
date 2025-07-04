import '@fontsource/roboto/300.css';
import '@fontsource/roboto/400.css';
import '@fontsource/roboto/500.css';
import '@fontsource/roboto/700.css';
import './App.css'
import CssBaseline from '@mui/material/CssBaseline';
import {Button, ButtonGroup, createTheme, Grid, ThemeProvider} from "@mui/material";
import {Log} from "./Log.jsx";
import {Macros} from "./Macros.jsx";

const theme = createTheme({
    palette: {
        mode: 'dark',  // 'light' or 'dark'
        primary: {
            main: '#1976d2',
        },
    },
});

function App() {
    return (
        <ThemeProvider theme={theme}>
            <CssBaseline/>
            <Grid container spacing={0} alignContent={'flex-start'} sx={{width: '100vw', height: '100vh'}}>
                <Grid size={6}><Macros /></Grid>
                <Grid size={6}><Log/>
                    <ButtonGroup variant="contained" sx={{gap: 4, display: 'flex', justifyContent: 'center'}}>
                        <Button color={'error'}>Stop</Button>
                        <Button>Start</Button>
                    </ButtonGroup>
                </Grid>
            </Grid>
        </ThemeProvider>
    )
}

export default App
