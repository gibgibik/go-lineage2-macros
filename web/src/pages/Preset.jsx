import React, {useContext, useEffect, useState} from "react";
import {Box, Button, Grid, ListItemButton, ListItemText} from "@mui/material";
import List from '@mui/material/List';
import {getPresetsList} from "../api.js";
import {NotificationContext} from "../components/Alert/NotificationContext.jsx";
import {Macros} from "../components/Macros/Macros.jsx";

const NEW_ID_NAME = 'New';

export const Preset = ({value, index, ...other}) => {
    if (value !== index) {
        return null;
    }
    const {setAlert} = useContext(NotificationContext);

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
    // useEffect(() => {
    //     console.log(presetsList);
    // }, [presetsList])
    useEffect(() => {
        if (!presetId) {
            return;
        }

    }, [presetId]);
    const handleListItemClick = (event, index) => {
        setPresetId(index);
    };

    const addNew = () => {
        if (Object.keys(presetsList).find((idx) => presetsList[idx].name === NEW_ID_NAME)) {
            return;
        }
        const now = Date.now();
        setPresetsList({...presetsList, [now]: {name: NEW_ID_NAME, id: now}});
        setPresetId(now);
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
                        <Macros presetId={presetId} loadPresets={loadPresets} presetName={presetsList[presetId].name} data={presetsList[presetId]} setPreset={(newPreset) => {
                            setPresetsList({
                                ...presetsList,
                                [presetId.toString()]: {...presetsList[presetId].id, name: newPreset.target.value}
                            });
                            return true;
                        }}/>}
                </Grid>
            </Grid>
        </Box>
    );
}
