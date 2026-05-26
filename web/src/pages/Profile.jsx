import React, {useCallback, useContext, useEffect, useMemo, useState} from "react";
import {Box, Button, Grid, ListItemButton, ListItemText} from "@mui/material";
import List from '@mui/material/List';
import {getPresetsList, getProfilesList, saveProfile} from "../api.js";
import {NotificationContext} from "../components/Alert/NotificationContext.jsx";
import {Macros} from "../components/Macros/Macros.jsx";
import {ProfilePreset} from "../components/ProfilePreset/ProfilePreset.jsx";
import {ProfileMacros} from "../components/ProfileMacros/ProfileMacros.jsx";
import {ProfileTarget} from "../components/ProfileTarget/ProfileTarget.jsx";

const NEW_PROFILE_NAME = 'New';

export const Profile = ({value, index, profileName, setProfileName, currentPid, ...other}) => {
    const {setAlert, setSuccess} = useContext(NotificationContext);

    const [profiles, setProfiles] = useState({});
    const [activePreset, setActivePreset] = useState(null);
    const [presetsList, setPresetsList] = useState([]);
    const [preferredTargets, setPreferredTargets] = useState([]);
    const [allowedTargets, setAllowedTargets] = useState([]);
    const loadData = () => {
        const fetchProfiles = async () => {
            try {
                const {data} = await getProfilesList();
                if (data) {
                    setProfiles(data.reduce((acc, item) => {
                        acc[item.name] = item;
                        return acc;
                    }, {}));
                    console.log(data);
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
    useEffect(() => {
        if (!profileName) {
            return;
        }
        if (Array.isArray(profiles[profileName].preferred_targets)) {
            setPreferredTargets(profiles[profileName].preferred_targets.map((item) => {
                return {title: item, value: item};
            }));
        }
    }, [profileName]);
    useEffect(() => {
        if (!profileName) {
            return;
        }
        if (Array.isArray(profiles[profileName].allowed_targets)) {
            setAllowedTargets(profiles[profileName].allowed_targets.map((item) => {
                return {title: item, value: item};
            }));
        }
    }, [profileName]);
    const addNew = () => {
        const value = prompt('Enter profile name');
        if (typeof profiles[value] !== 'undefined') {
            return;
        }
        setProfiles({...profiles, [value]: {name: value, items: []}});
        setProfileName(value);
    }
    const save = useCallback(formData => {
        const save = () => {
            try {
                 saveProfile(profileName, currentPid, {
                    ...profiles[profileName],
                    preferred_targets: preferredTargets.map((item) => item.value),
                    allowed_targets: allowedTargets.map((item) => item.value),
                    items: profiles[profileName].items.map((item) => {
                        if (item.preset.name == formData.name) {
                            item.preset = formData;
                            return item;
                        } else {
                            return item;
                        }
                    })
                });
                setSuccess('Saved');
            } catch (error) {
                setAlert(error.message);
            }
        }
        save();
    }, [allowedTargets, currentPid, preferredTargets, profileName, profiles, setAlert, setSuccess]);
    const isBatchRun = useMemo(() => {
        if (!profileName) {
            return false;
        }
        return !!profiles[profileName].items.find(item => item.preset.id == activePreset && item.batch_run === true);
    }, [activePreset, profiles]);
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
                    <Box>
                        {profileName &&
                            <ProfileTarget preferredTargets={preferredTargets} setPreferredTargets={setPreferredTargets}
                                           allowedTargets={allowedTargets} setAllowedTargets={setAllowedTargets}/>}
                    </Box>
                    {profileName &&
                        <ProfilePreset data={profiles[profileName]} setProfiles={setProfiles} profiles={profiles}
                                       setActivePreset={setActivePreset} presetsList={presetsList}/>}
                </Grid>
                <Grid size={7}>
                    {activePreset > 0 &&
                        <ProfileMacros presetId={activePreset} profiles={profiles} setProfiles={setProfiles}
                                       profileName={profileName} onSave={save}
                                       presetName={presetsList[activePreset].name}
                                       data={profiles[profileName]?.items.find(item => item.preset.id == activePreset)?.preset}
                                       isBatchRun={isBatchRun}/>}
                </Grid>
            </Grid>
        </Box>
    );
}
