CREATE TYPE task_status AS ENUM ('new', 'in_progress', 'completed', 'error');
CREATE TYPE gender AS ENUM ('male', 'female');

CREATE TABLE IF NOT EXISTS task (
    id SERIAL PRIMARY KEY,
    task_status task_status NOT NULL DEFAULT 'new',

    faces_total INT NOT NULL DEFAULT 0,
    faces_female INT NOT NULL DEFAULT 0,
    faces_male INT NOT NULL DEFAULT 0,

    age_female_avg INT NOT NULL DEFAULT 0,
    age_male_avg INT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS task_image (
    id SERIAL PRIMARY KEY,
    task_id INT NOT NULL, 

    image_name TEXT NOT NULL,
    done boolean NOT NULL DEFAULT false,

    FOREIGN KEY (task_id) REFERENCES task (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS face (
    id SERIAL PRIMARY KEY,
    image_id INT NOT NULL, 

    gender gender NOT NULL,
    age SMALLINT NOT NULL,
    
    bbox_height INT NOT NULL,
    bbox_width INT NOT NULL,
    bbox_x INT NOT NULL,
    bbox_y INT NOT NULL,

    FOREIGN KEY (image_id) REFERENCES task_image (id) ON DELETE CASCADE
);

