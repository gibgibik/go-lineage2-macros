import React, {useContext, useEffect, useState} from "react";
import {Box, Button, Grid, ListItemButton, ListItemText} from "@mui/material";
import List from '@mui/material/List';
import {getPresetsList, savePreset} from "../api.js";
import {NotificationContext} from "../components/Alert/NotificationContext.jsx";
import {Macros} from "../components/Macros/Macros.jsx";

export const Preset = ({value, index, ...other}) => {
    const {setAlert, setSuccess} = useContext(NotificationContext);

    const [presetId, setPresetId] = useState(null);
    const [presetsList, setPresetsList] = useState({});
    const loadPresets = () => {
        const fetchPresets = async () => {
            try {
                const {data} = await getPresetsList();
                setPresetsList(data.reduce((acc, item) => {
                    acc[item.id] = item;
                    return acc;
                }, {}));
            } catch (error) {
                setAlert(error.response?.data);
            }
        }
        fetchPresets();
    }
    useEffect(() => loadPresets(), []);
    useEffect(() => {
        if (!presetId) {
            return;
        }

    }, [presetId]);
    const handleListItemClick = (event, index) => {
        setPresetId(index);
    };

    const addNew = () => {
        const val = prompt("Enter New Preset");
        if (Object.keys(presetsList).find((idx) => presetsList[idx].name === val)) {
            return;
        }
        const now = Date.now();
        setPresetsList({...presetsList, [now]: {name: val, id: now}});
        setPresetId(now);
    }
    const macrosSave = (presetData) => {
        const save = async () => {
            try {
                await savePreset(presetId, {...presetData, id: presetId});
            } catch (error) {
                setAlert(error.message);
                return;
            }
            setSuccess('Saved');
            await loadPresets();
        };
        save();
    }

    if (value !== index) {
        return null;
    }

    return (
        <Box
            role="tabpanel"
            hidden={value !== index}
            id={`simple-tabpanel-${index}`}
            aria-labelledby={`simple-tab-${index}`}
            {...other}
        >
            <Grid container sx={{width: '100%'}} spacing={2}>
                <Grid size={3} sx={{borderRight: '1px solid #ddd'}}>
                    <List sx={{
                        width: '100%',
                    }}>
                        {Object.keys(presetsList).map((pId) => {
                            return (<ListItemButton href="#simple-list" selected={presetId == pId}
                                                    key={pId}
                                                    onClick={(event) => handleListItemClick(event, presetsList[pId].id)}
                                                    sx={{width: '100%'}}>
                                <ListItemText primary={presetsList[pId].name}/>
                            </ListItemButton>);
                        })}
                    </List>
                    <Grid sx={{paddingLeft: 2, paddingRight: 2}} container alignItems={'center'}
                          justifyContent={'center'} spacing={2}>
                        {/*<TextField value={""} style={{width:'100%'}}/>*/}
                        <Button variant={"contained"} onClick={addNew}>Add</Button>
                    </Grid>
                </Grid>
                <Grid size={9}>
                    {presetId &&
                        <Macros presetId={presetId} onSave={macrosSave} presetName={presetsList[presetId].name} data={presetsList[presetId]} />}
                </Grid>
            </Grid>
        </Box>
    );
}
