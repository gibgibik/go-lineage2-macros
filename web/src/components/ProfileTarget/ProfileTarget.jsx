import {Box, Checkbox, TextField} from "@mui/material";
import Autocomplete from '@mui/material/Autocomplete';
import CheckBoxOutlineBlankIcon from '@mui/icons-material/CheckBoxOutlineBlank';
import CheckBoxIcon from '@mui/icons-material/CheckBox';
import {useEffect, useState} from "react";
import {getNpcList} from "../../api.js";


// Top 100 films as rated by IMDb users. http://www.imdb.com/chart/top
const top100Films = [
    {title: 'The Shawshank Redemption', value: 'asd'}
];

const top101Films = [
    {title: 'The Shawshank Redemption', value: 'asd'}
];

const icon = <CheckBoxOutlineBlankIcon fontSize="small"/>;
const checkedIcon = <CheckBoxIcon fontSize="small"/>;

export const ProfileTarget = ({preferredTargets, setPreferredTargets, allowedTargets, setAllowedTargets}) => {
    const [preferredNpcList, setPreferredNpcList] = useState([]);
    const [allowedNpcList, setAllowedNpcList] = useState([]);
    useEffect(() => {
        const f = async () => {
            const {data} = await getNpcList();
            const prepared = Object.keys(data).map((key) => {
                return {title: key, value: key,}
            });
            setPreferredNpcList(prepared);
            setAllowedNpcList(prepared);
        }
        f();
    }, []);
    if (!preferredNpcList.length || !allowedNpcList.length) {
        return;
    }
    return <Box sx={{m: 2}}><Autocomplete
        sx={{mb: 2}}
        multiple
        options={preferredNpcList}
        value={preferredTargets}
        disableCloseOnSelect
        getOptionLabel={(option) => option.title}
        onChange={(props, newValue) => {
            setPreferredTargets(newValue);
        }}
        renderOption={(props, option, {selected}) => {
            const {key, ...optionProps} = props;
            return (
                <li key={key} {...optionProps}>
                    <Checkbox
                        icon={icon}
                        checkedIcon={checkedIcon}
                        style={{marginRight: 8}}
                        checked={selected}
                    />
                    {option.title}
                </li>
            );
        }}
        renderInput={(params) => (
            <TextField {...params} label="Preferred Targets(AI)"/>
        )}
    />
        <Autocomplete
            multiple
            value={allowedTargets}
            options={allowedNpcList}
            disableCloseOnSelect
            getOptionLabel={(option) => option.title}
            onChange={(props, newValue) => {
                setAllowedTargets(newValue);
            }}
            renderOption={(props, option, {selected}) => {
                const {key, ...optionProps} = props;
                return (
                    <li key={key} {...optionProps}>
                        <Checkbox
                            icon={icon}
                            checkedIcon={checkedIcon}
                            style={{marginRight: 8}}
                            checked={selected}
                        />
                        {option.title}
                    </li>
                );
            }}
            renderInput={(params) => (
                <TextField {...params} label="Allowed Targets"/>
            )}
        />
    </Box>;
}
