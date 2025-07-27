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
import React, {useEffect, useState} from "react";


export const ProfilePreset = ({data, setProfiles, profiles, setActivePreset, presetsList}) => {
    const [presetValue, setPresetValue] = useState('');
    const [chosenPresetList, setChosenPresetList] = useState({});
    const [chosenPreset, setChosenPreset] = useState(null);
    const addNew = () => {
        if (typeof (chosenPresetList?.[presetValue]) !== 'undefined') {
            return
        }
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
    useEffect(() => {
        if (!data) {
            return;
        }
        setChosenPresetList({...chosenPresetList, ...data.items.reduce((acc, cur) => {acc[cur.preset.id] = cur.preset; return acc}, {})});
    }, [data]);
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
