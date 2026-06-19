from sqlalchemy import Column, Integer, String
from database import Base
import numpy as np

class User(Base):
    __tablename__ = 'users'
    
    id = Column(Integer, primary_key=True)
    name = Column(String)
    email = Column(String)
    
    def __repr__(self):
        return f"<User(id={self.id}, name={self.name})>"

def calculate_stats(values):
    arr = np.array(values)
    return {
        'mean': float(np.mean(arr)),
        'std': float(np.std(arr)),
        'min': float(np.min(arr)),
        'max': float(np.max(arr))
    }
