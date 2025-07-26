import React, {useContext, useEffect, useState} from "react";
import {Box, Button, Grid, ListItemButton, ListItemText} from "@mui/material";
import List from '@mui/material/List';
import {getPresetsList, getProfilesList} from "../api.js";
import {NotificationContext} from "../components/Alert/NotificationContext.jsx";
import {Macros} from "../components/Macros/Macros.jsx";
import {ProfilePreset} from "../components/ProfilePreset/ProfilePreset.jsx";

const NEW_PROFILE_NAME = 'New';

export const Profile = ({value, index, ...other}) => {
    if (value !== index) {
        return null;
    }
    const {setAlert} = useContext(NotificationContext);

    const [profileName, setProfileName] = useState(null);
    const [profiles, setProfiles] = useState({});
    const [activePreset, setActivePreset] = useState(null);
    const [presetsList, setPresetsList] = useState([]);
    const loadData = () => {
        const fetchProfiles = async () => {
            try {
                const {data} = await getProfilesList();
                if (data) {
                    setProfiles(data.reduce((acc, item) => {
                        acc[item.id] = item;
                        return acc;
                    }, {}));
                }

            } catch (error) {
                setAlert(error.response?.data);
            }
        }
        fetchProfiles();
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
    useEffect(() => loadData(), []);
    const handleListItemClick = (profileName) => {
        setProfileName(profileName);
    };

    const addNew = () => {
        const value = prompt('Enter profile name');
        if (typeof profiles[value] !== 'undefined') {
            return;
        }
        setProfiles({...profiles, [value]: {name: value}});
        setProfileName(value);
    }
    console.log(activePreset);
    return (
        <Box
            role="tabpanel"
            hidden={value !== index}
            id={`simple-tabpanel-${index}`}
            aria-labelledby={`simple-tab-${index}`}
            {...other}
        >
            <Grid container sx={{width: '100%'}}>
                <Grid size={3} sx={{borderRight: '1px solid #ddd'}}>
                    <List sx={{
                        width: '100%',
                    }}>
                        {Object.keys(profiles).map((pId) => {
                            return (<ListItemButton href="#simple-list" selected={profileName == pId}
                                                    key={pId}
                                                    onClick={(event) => handleListItemClick(pId)}
                                                    sx={{width: '100%'}}>
                                <ListItemText primary={profiles[pId].name}/>
                            </ListItemButton>);
                        })}
                    </List>
                    <Grid sx={{paddingLeft: 2, paddingRight: 2}} container alignItems={'center'}
                          justifyContent={'center'} spacing={2}>
                        {/*<TextField value={""} style={{width:'100%'}}/>*/}
                        <Button variant={"contained"} onClick={addNew}>Add</Button>
                    </Grid>
                </Grid>
                <Grid size={2} sx={{borderRight: '1px solid #ddd'}}>
                    {profileName &&
                        <ProfilePreset data={profiles[profileName]} setProfiles={setProfiles} profiles={profiles}
                                       setActivePreset={setActivePreset} presetsList={presetsList} setPresetsList={setPresetsList}/>}
                </Grid>
                <Grid size={7}>
                    {activePreset &&
                        <Macros presetId={activePreset} loadPresets={() => {
                            console.log('todo save')
                        }} presetName={presetsList[activePreset].name} data={presetsList[activePreset]}
                                setPreset={(newPreset) => {
                                    setPresetsList({
                                        ...presetsList,
                                        [presetId.toString()]: {
                                            ...presetsList[activePreset].id,
                                            name: newPreset.target.value
                                        }
                                    });
                                    return true;
                                }}/>}
                </Grid>
            </Grid>
        </Box>
    );
}
