import List from '@mui/material/List';
import {
    Box,
    Button,
    Checkbox,
    FormControl,
    Grid,
    InputLabel,
    ListItemButton,
    ListItemText,
    MenuItem,
    Select, Typography
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

        setProfiles({
            ...profiles,
            [data.name]: {
                ...profiles[data.name],
                items: [...profiles[data.name].items || [], {
                    is_active: true,
                    batch_run: false,
                    preset: presetsList[presetValue]
                }],
                is_active: true,
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
    const isActiveChangeHandler = (event, pId) => {
        setProfiles({
            ...profiles,
            [data.name]: {
                ...profiles[data.name],
                items: [...(profiles[data.name].items || []).map((item) => {
                    if (item.preset.id == pId) {
                        item.is_active = event.target.checked;
                    }
                    return item;
                })],
            }
        });
    }
    useEffect(() => {
        if (!data?.items) {
            return;
        }
        setChosenPresetList(data.items.reduce((acc, cur) => {
                acc[cur.preset.id] = cur.preset;
                acc[cur.preset.id].batch_run = cur.batch_run;
                acc[cur.preset.id].is_active = cur.is_active;
                return acc
            }, {})
        );
    }, [data]);
    return <Box sx={{m: 2}}>
        <Typography variant={'h6'}>Presets</Typography>
        <List sx={{width: '100%'}}>
            {Object.keys(chosenPresetList).map((pId) => {
                return (<Box key={pId}>
                    <ListItemButton href="#simple-list" selected={chosenPreset == pId}
                                    key={pId}
                                    onClick={(event) => handlePresetChange(pId)}>
                        <Checkbox onChange={(e) => isActiveChangeHandler(e, pId)}
                                  checked={chosenPresetList[pId]?.is_active}/>
                        <ListItemText primary={chosenPresetList[pId].name}/>
                    </ListItemButton></Box>);
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
