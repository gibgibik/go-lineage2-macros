import {Box, Container, FormControl, InputLabel, MenuItem, Select, TextField} from "@mui/material";

const rawActions = [
    '/assist',
    '/attack',
    '/target',
    '/delay',
    '/useskill',
    '/press',
];
const actions = <Select
    variant={'outlined'}
    id="demo-simple-select"
    label="Age"
>{rawActions.map((item, k) => <MenuItem key={k} value={item}>{item}</MenuItem>)}</Select>;
const actionField = <FormControl fullWidth={true}>
    <InputLabel id="demo-simple-select-label">Action</InputLabel>
    {actions}
</FormControl>;
const items = []
for (let i = 0; i < 10; i++) {
    items.push(<Box sx={{display: 'flex', gap: 2, m: 2}}>
        {actionField}
        <TextField variant={"outlined"} fullWidth={true}>sd</TextField>
    </Box>);
}
export const Macros = () => {
    return (<Container>
        {items}

    </Container>)
        ;
}
