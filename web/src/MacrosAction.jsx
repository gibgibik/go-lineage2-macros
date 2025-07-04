import {FormControl, InputLabel, MenuItem, Select} from "@mui/material";

const rawActions = [
    '/assist',
    '/attack',
    '/target',
    '/delay',
    '/useskill',
    '/press',
];
const actions = rawActions.map((item, k) => <MenuItem key={k} value={item}>{item}</MenuItem>);
export const MacrosAction = (props) => {
    return <FormControl fullWidth={true} >
        <InputLabel id="action-select-label">Action</InputLabel>
        <Select
            variant={'outlined'}
            id="action-simple-select"
            label="Action"
            {...props}
        >
            {actions}

        </Select>
    </FormControl>;
}
