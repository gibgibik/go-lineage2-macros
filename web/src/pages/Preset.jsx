import React, {useContext, useEffect, useState} from "react";
import {Box, Button, Grid, ListItemButton, ListItemText, TextField} from "@mui/material";
import List from '@mui/material/List';
import {getPresetsList} from "../api.js";
import {NotificationContext} from "../components/Alert/NotificationContext.jsx";

export const Preset = ({value, index, ...other}) => {
    if (value !== index) {
        return null;
    }
    const {setAlert, setSuccess} = useContext(NotificationContext);

    const [selectedIndex, setSelectedIndex] = useState(1);
    const [presetsList, setPresetsList] = useState([]);
    useEffect(() => {
        const fetchPresets = async () => {
            const presets = getPresetsList();

        }
        console.log('first load');
    }, []);
    const handleListItemClick = (event, index) => {
        setSelectedIndex(index);
    };

    return (
        <Box
            role="tabpanel"
            hidden={value !== index}
            id={`simple-tabpanel-${index}`}
            aria-labelledby={`simple-tab-${index}`}
            {...other}
        >
            <Grid container sx={{width: '100%'}} spacing={2}>
                <Grid size={2} sx={{borderRight: '1px solid #ddd'}}>
                    <List sx={{
                        width: '100%',
                    }}>
                        <ListItemButton component="a" href="#simple-list" selected={selectedIndex === 0}
                                        onClick={(event) => handleListItemClick(event, 0)} sx={{width: '100%'}}>
                            <ListItemText primary="Spam"/>
                        </ListItemButton>
                        <ListItemButton component="a" href="#simple-list" selected={selectedIndex === 1}
                                        onClick={(event) => handleListItemClick(event, 1)}>
                            <ListItemText primary="Spam"/>
                        </ListItemButton>
                    </List>
                    <Grid sx={{paddingLeft: 2, paddingRight: 2}} container alignItems={'center'} justifyContent={'center'} spacing={2}>
                        <TextField value={""} style={{width:'100%'}}/>
                        <Button variant={"contained"}>Add</Button>
                    </Grid>
                </Grid>
                <Grid size={10} >
                    as
                </Grid>
            </Grid>
        </Box>
    );
}
