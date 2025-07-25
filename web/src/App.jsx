import '@fontsource/roboto/300.css';
import '@fontsource/roboto/400.css';
import '@fontsource/roboto/500.css';
import '@fontsource/roboto/700.css';
import './App.css'
import CssBaseline from '@mui/material/CssBaseline';
import {Box, createTheme, Tab, Tabs, ThemeProvider} from "@mui/material";
import React from "react";
import {Preset} from "./pages/Preset.jsx";
import {Running} from "./pages/Running.jsx";
import {Profile} from "./pages/Profile.jsx";
import {NotificationContext, AlertProvider} from "./components/Alert/NotificationContext.jsx";

const theme = createTheme({
    palette: {
        mode: 'dark',  // 'light' or 'dark'
        primary: {
            main: '#1976d2',
        },
    },
});

function App() {
    const [currentTab, setTab] = React.useState(0);

    const handleChangeTab = (event, newValue) => {
        setTab(newValue);
    };

    return (
        <ThemeProvider theme={theme}>
            <CssBaseline/>
            <AlertProvider>
                <Box sx={{borderBottom: 1, borderColor: 'divider'}}>
                    <Tabs value={currentTab} onChange={handleChangeTab} aria-label="basic tabs example">
                        <Tab label="Running"/>
                        <Tab label="Profiles"/>
                        <Tab label="Presets"/>
                    </Tabs>
                </Box>
                <Running value={currentTab} index={0}/>
                <Profile value={currentTab} index={1}/>
                <Preset value={currentTab} index={2}/>
            </AlertProvider>
        </ThemeProvider>
    )
}

export default App
