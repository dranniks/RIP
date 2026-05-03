import Button from "react-bootstrap/Button";
import Form from "react-bootstrap/Form";

export interface ServiceFilterState {
  query: string;
}

interface ServicesFiltersProps {
  value: ServiceFilterState;
  onChange: (next: ServiceFilterState) => void;
  onSubmit: () => void;
}

export const ServicesFilters = ({ value, onChange, onSubmit }: ServicesFiltersProps): JSX.Element => {
  const handleSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
    event.preventDefault();
    onSubmit();
  };

  return (
    <Form className="search-card" onSubmit={handleSubmit}>
      <div className="search-row">
        <Form.Control
          id="q"
          className="search-input"
          type="text"
          placeholder="Поиск по названию услуги"
          value={value.query}
          onChange={(event) =>
            onChange({
              query: event.target.value
            })
          }
        />
        <Button variant="link" className="chip chip-button search-submit" type="submit">
          Поиск
        </Button>
      </div>
    </Form>
  );
};
