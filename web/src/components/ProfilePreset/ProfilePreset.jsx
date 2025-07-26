import List from '@mui/material/List';
import {
    Box,
    Button,
    FormControl,
    Grid,
    InputLabel,
    ListItemButton,
    ListItemText,
    MenuItem,
    Select
} from "@mui/material";
import React, {useContext, useEffect, useState} from "react";
import {getPresetsList} from "../../api.js";
import {NotificationContext} from "../Alert/NotificationContext.jsx";


export const ProfilePreset = ({data, setProfiles, profiles, setActivePreset, presetsList, setPresetsList}) => {
    useEffect(() => {
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
    }, []);
    const [presetValue, setPresetValue] = useState('');
    const [chosenPresetList, setChosenPresetList] = useState({});
    const [chosenPreset, setChosenPreset] = useState(null);
    const {setAlert} = useContext(NotificationContext);
    const addNew = () => {
        setChosenPresetList({...chosenPresetList, [presetValue]: presetsList[presetValue]});
        setProfiles({...profiles,
            [data.name]: {
                ...profiles[data.name],
                items: [...profiles[data.name].items || [], {is_active: true, preset: presetsList[presetValue]}]
            }
        });
    }
    const handleChange = (event) => {
        setPresetValue(event.target.value);
    };
    const handlePresetChange = (val) => {
        setChosenPreset(val);
        setActivePreset(val);
    };
    return <Box>
        <List sx={{width: '100%'}}>
            {Object.keys(chosenPresetList).map((pId) => {
                return (<ListItemButton href="#simple-list" selected={chosenPreset == pId}
                                        key={pId}
                                        onClick={(event) => handlePresetChange(pId)}
                                        sx={{width: '100%'}}>
                    <ListItemText primary={chosenPresetList[pId].name}/>
                </ListItemButton>);
            })}
        </List>
        <Grid sx={{paddingLeft: 2, paddingRight: 2}} container alignItems={'center'} flexDirection={'row'}
              justifyContent={'center'} spacing={2}>
            <FormControl sx={{flexGrow: 1}}>
                <InputLabel id={'preset-label'}/>
                <Select labelId={'preset-label'} value={presetValue} name="preset" variant={'filled'}
                        onChange={handleChange}>
                    {Object.keys(presetsList).map(key => <MenuItem value={key}>{presetsList[key].name}</MenuItem>)}
                </Select>
            </FormControl>
            <Button variant={"contained"} onClick={addNew}>Add</Button>
        </Grid>
    </Box>;
}
